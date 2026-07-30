package stepfun

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const (
	endpointProfileField = "endpoint_profile"
	defaultSpeechURL     = "https://api.stepfun.com/v1/audio/speech"
	stepPlanSpeechURL    = "https://api.stepfun.com/step_plan/v1/audio/speech"
)

type speechRequest struct {
	Model            string         `json:"model"`
	Input            string         `json:"input"`
	Voice            string         `json:"voice"`
	ResponseFormat   string         `json:"response_format,omitempty"`
	Speed            float64        `json:"speed,omitempty"`
	Volume           *float64       `json:"volume,omitempty"`
	VoiceLabel       map[string]any `json:"voice_label,omitempty"`
	Instruction      string         `json:"instruction,omitempty"`
	SampleRate       *int           `json:"sample_rate,omitempty"`
	PronunciationMap map[string]any `json:"pronunciation_map,omitempty"`
	StreamFormat     string         `json:"stream_format,omitempty"`
	MarkdownFilter   *bool          `json:"markdown_filter,omitempty"`
}

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opt := options.MustGet()

	speechEndpoint, endpointErr := resolveSpeechEndpoint(opt.ExtraBody)
	if endpointErr != nil {
		return mo.Err[any](endpointErr)
	}

	values, buildErr := buildSpeechRequest(opt)
	if buildErr != nil {
		return mo.Err[any](buildErr)
	}

	payload := lo.Must(json.Marshal(values))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		speechEndpoint,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](
			apierrors.NewErrBadGateway().
				WithDetail(err.Error()).
				WithError(err).
				WithCaller(),
		)
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= http.StatusBadRequest {
		ct := res.Header.Get("Content-Type")

		switch {
		case strings.HasPrefix(ct, "application/json"):
			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
			)
		case strings.HasPrefix(ct, "text/"):
			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail(utils.NewTextResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
			)
		default:
			slog.Warn("unknown upstream error",
				slog.Int("status", res.StatusCode),
				slog.String("content_type", ct),
				slog.String("content_length", res.Header.Get("Content-Length")),
			)

			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail("unknown Content-Type: " + ct),
			)
		}
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		contentType = contentTypeForFormat(opt.ResponseFormat)
	}

	return mo.Ok[any](c.Stream(http.StatusOK, contentType, res.Body))
}

func resolveSpeechEndpoint(extraBody map[string]any) (string, error) {
	// A profile is a provider-owned endpoint identity, not a caller-supplied
	// URL. Keeping the URL mapping here preserves StepFun as the single source
	// of truth and prevents this proxy from becoming an SSRF primitive.
	switch utils.GetByJSONPath[string](extraBody, "{ ."+endpointProfileField+" }") {
	case "", "default":
		return defaultSpeechURL, nil
	case "step-plan":
		return stepPlanSpeechURL, nil
	default:
		return "", apierrors.NewErrBadRequest().
			WithDetail("unsupported stepfun endpoint profile")
	}
}

func buildSpeechRequest(opt types.SpeechRequestOptions) (speechRequest, error) {
	body := speechRequest{
		Model:          opt.Model,
		Input:          opt.Input,
		Voice:          opt.Voice,
		ResponseFormat: opt.ResponseFormat,
		Speed:          opt.Speed,
	}
	if len(opt.Input) > 1000 {
		return body, apierrors.NewErrBadRequest().WithDetail("stepfun tts input must be at most 1000 characters")
	}

	extra := opt.ExtraBody
	if volume := utils.GetByJSONPath[*float64](extra, "{ .volume }"); volume != nil {
		body.Volume = volume
	}
	if voiceLabel := utils.GetByJSONPath[map[string]any](extra, "{ .voice_label }"); voiceLabel != nil {
		if opt.Model == modelStepAudio25TTS {
			return body, apierrors.NewErrBadRequest().WithDetail("stepaudio-2.5-tts does not support voice_label; use instruction or inline parentheses prompts")
		}
		body.VoiceLabel = voiceLabel
	}
	if instruction := utils.GetByJSONPath[string](extra, "{ .instruction }"); instruction != "" {
		if opt.Model != modelStepAudio25TTS {
			return body, apierrors.NewErrBadRequest().WithDetail("stepfun instruction is only supported by stepaudio-2.5-tts")
		}
		body.Instruction = instruction
	}
	if sampleRate := utils.GetByJSONPath[*int](extra, "{ .sample_rate }"); sampleRate != nil {
		body.SampleRate = sampleRate
	}
	if pronunciationMap := utils.GetByJSONPath[map[string]any](extra, "{ .pronunciation_map }"); pronunciationMap != nil {
		body.PronunciationMap = pronunciationMap
	}
	if streamFormat := utils.GetByJSONPath[string](extra, "{ .stream_format }"); streamFormat != "" {
		body.StreamFormat = streamFormat
	}
	if markdownFilter := utils.GetByJSONPath[*bool](extra, "{ .markdown_filter }"); markdownFilter != nil {
		body.MarkdownFilter = markdownFilter
	}

	return body, nil
}

func contentTypeForFormat(format string) string {
	switch format {
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	case "opus":
		return "audio/opus"
	case "pcm":
		return "audio/L16"
	default:
		return "audio/mpeg"
	}
}
