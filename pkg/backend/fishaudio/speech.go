package fishaudio

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

const defaultSpeechURL = "https://api.fish.audio/v1/tts"

// prosody carries Fish Audio's speech pacing controls.
//
// https://docs.fish.audio/api-reference/endpoint/openapi-v1/text-to-speech
type prosody struct {
	// Speed multiplier, 1.0 is normal speed.
	Speed float64 `json:"speed,omitempty"`
	// Volume gain in dB, 0 is unchanged.
	Volume float64 `json:"volume,omitempty"`
}

// speechRequest is the Fish Audio TTS request payload.
//
// https://docs.fish.audio/api-reference/endpoint/openapi-v1/text-to-speech
type speechRequest struct {
	Text string `json:"text"`
	// ReferenceID selects the voice model (the id in a fish.audio voice page URL).
	ReferenceID string   `json:"reference_id,omitempty"`
	Format      string   `json:"format,omitempty"`
	MP3Bitrate  *int     `json:"mp3_bitrate,omitempty"`
	ChunkLength *int     `json:"chunk_length,omitempty"`
	Normalize   *bool    `json:"normalize,omitempty"`
	Latency     string   `json:"latency,omitempty"`
	Prosody     *prosody `json:"prosody,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	SampleRate  *int     `json:"sample_rate,omitempty"`
}

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opt := options.MustGet()

	body, buildErr := buildSpeechRequest(opt)
	if buildErr != nil {
		return mo.Err[any](buildErr)
	}

	payload := lo.Must(json.Marshal(body))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		defaultSpeechURL,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
	req.Header.Set("Content-Type", "application/json")
	// NOTICE: Fish Audio selects the TTS model generation (s1, s2-pro, ...) via a
	// non-standard `model` HTTP header, not the request body.
	// https://docs.fish.audio/api-reference/endpoint/openapi-v1/text-to-speech
	//nolint:canonicalheader
	req.Header.Set("model", modelForRequest(opt.Model))

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
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = contentTypeForFormat(opt.ResponseFormat)
	}

	return mo.Ok[any](c.Stream(http.StatusOK, contentType, res.Body))
}

// modelForRequest resolves the `model` header value. An empty model (request
// sent as bare "fishaudio" without a "/<model>" suffix) falls back to s1 so
// the header is always explicit instead of relying on upstream defaults.
func modelForRequest(model string) string {
	if model == "" {
		return modelS1
	}

	return model
}

func buildSpeechRequest(opt types.SpeechRequestOptions) (speechRequest, error) {
	body := speechRequest{
		Text: opt.Input,
		// OpenAI `voice` maps to Fish Audio's reference voice model id.
		ReferenceID: opt.Voice,
	}

	format, err := formatForResponseFormat(opt.ResponseFormat)
	if err != nil {
		return body, err
	}

	body.Format = format

	// OpenAI `speed` maps to prosody.speed; an explicit prosody object in
	// extra_body takes precedence since it is the provider-native control.
	if opt.Speed != 0 {
		body.Prosody = &prosody{Speed: opt.Speed}
	}

	extra := opt.ExtraBody
	if prosodyMap := utils.GetByJSONPath[map[string]any](extra, "{ .prosody }"); prosodyMap != nil {
		extraProsody := &prosody{}
		if speed, ok := prosodyMap["speed"].(float64); ok {
			extraProsody.Speed = speed
		}

		if volume, ok := prosodyMap["volume"].(float64); ok {
			extraProsody.Volume = volume
		}

		body.Prosody = extraProsody
	}

	if chunkLength := utils.GetByJSONPath[*int](extra, "{ .chunk_length }"); chunkLength != nil {
		body.ChunkLength = chunkLength
	}

	if normalize := utils.GetByJSONPath[*bool](extra, "{ .normalize }"); normalize != nil {
		body.Normalize = normalize
	}

	if latency := utils.GetByJSONPath[string](extra, "{ .latency }"); latency != "" {
		body.Latency = latency
	}

	if temperature := utils.GetByJSONPath[*float64](extra, "{ .temperature }"); temperature != nil {
		body.Temperature = temperature
	}

	if topP := utils.GetByJSONPath[*float64](extra, "{ .top_p }"); topP != nil {
		body.TopP = topP
	}

	if sampleRate := utils.GetByJSONPath[*int](extra, "{ .sample_rate }"); sampleRate != nil {
		body.SampleRate = sampleRate
	}

	if mp3Bitrate := utils.GetByJSONPath[*int](extra, "{ .mp3_bitrate }"); mp3Bitrate != nil {
		body.MP3Bitrate = mp3Bitrate
	}

	return body, nil
}

// formatForResponseFormat maps OpenAI response_format values onto Fish Audio
// formats. Fish Audio supports mp3, wav, pcm, and opus; aac and flac have no
// upstream equivalent and are rejected instead of being silently transcoded.
func formatForResponseFormat(responseFormat string) (string, error) {
	switch responseFormat {
	case "", "mp3":
		return "mp3", nil
	case "wav", "pcm", "opus":
		return responseFormat, nil
	default:
		return "", apierrors.NewErrBadRequest().
			WithDetail("fish audio supports only mp3, wav, pcm, and opus response formats")
	}
}

func contentTypeForFormat(format string) string {
	switch format {
	case "wav":
		return "audio/wav"
	case "opus":
		return "audio/opus"
	case "pcm":
		return "audio/L16"
	default:
		return "audio/mpeg"
	}
}
