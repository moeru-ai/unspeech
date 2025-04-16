package volcengine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type SpeechRequestOptionsApp struct {
	AppID   string `json:"appid"`
	Token   string `json:"token"`
	Cluster string `json:"cluster"`
}

type SpeechRequestOptionsUser struct {
	UserID string `json:"uid"`
}

type SpeechRequestOptionsAudio struct {
	VoiceType        string   `json:"voice_type"`
	Emotion          *string  `json:"emotion,omitempty"`
	EnableEmotion    *bool    `json:"enable_emotion,omitempty"`
	EmotionScale     *float64 `json:"emotion_scale,omitempty"`
	Encoding         *string  `json:"encoding,omitempty"`
	SpeedRatio       *float64 `json:"speed_ratio,omitempty"`
	Rate             *int     `json:"rate,omitempty"`
	BitRate          *int     `json:"bit_rate,omitempty"`
	ExplicitLanguage *string  `json:"explicit_language,omitempty"`
	ContextLanguage  *string  `json:"context_language,omitempty"`
	LoudnessRatio    *float64 `json:"loudness_ratio,omitempty"`
}

type SpeechRequestOptionsRequest struct {
	RequestID             string         `json:"reqid"`
	Text                  string         `json:"text"`
	TextType              *string        `json:"text_type,omitempty"`
	SilenceDuration       *float64       `json:"silence_duration,omitempty"`
	WithTimestamp         *string        `json:"with_timestamp,omitempty"`
	Operation             *string        `json:"operation,omitempty"`
	ExtraParam            *string        `json:"extra_param,omitempty"`
	DisableMarkdownFilter *bool          `json:"disable_markdown_filter,omitempty"`
	EnableLatexTone       *bool          `json:"enable_latex_tn,omitempty"`
	CacheConfig           map[string]any `json:"cache_config,omitempty"`
	UseCache              *bool          `json:"use_cache,omitempty"`
}

type SpeechRequestOptions struct {
	App     SpeechRequestOptionsApp     `json:"app"`
	User    SpeechRequestOptionsUser    `json:"user"`
	Audio   SpeechRequestOptionsAudio   `json:"audio"`
	Request SpeechRequestOptionsRequest `json:"request"`
}

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opts := options.MustGet()

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	cluster := utils.GetByJSONPath[string](opts.ExtraBody, "{ .app.cluster }")
	if cluster == "" {
		cluster = "volcano_tts"
	}

	userID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .user.uid }")
	if userID == "" {
		userID = uuid.New().String()
	}

	requestID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .request.reqid }")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	operation := utils.GetByJSONPath[*string](opts.ExtraBody, "{ .request.operation }")
	if operation == nil || *operation == "" {
		operation = lo.ToPtr("query")
	}

	speedRatio := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.speed_ratio }")
	if speedRatio == nil || *speedRatio == 0 {
		speedRatio = lo.ToPtr(1.0)
	}

	newReqParams := &SpeechRequestOptions{
		App: SpeechRequestOptionsApp{
			AppID:   utils.GetByJSONPath[string](opts.ExtraBody, "{ .app.appid }"),
			Token:   token,
			Cluster: cluster,
		},
		User: SpeechRequestOptionsUser{
			UserID: userID,
		},
		Audio: SpeechRequestOptionsAudio{
			VoiceType:        opts.Voice,
			Emotion:          utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.emotion }"),
			EnableEmotion:    utils.GetByJSONPath[*bool](opts.ExtraBody, "{ .audio.enable_emotion }"),
			EmotionScale:     utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.emotion_scale }"),
			Encoding:         lo.Ternary(opts.ResponseFormat != "", lo.ToPtr(opts.ResponseFormat), lo.ToPtr("mp3")),
			SpeedRatio:       speedRatio,
			Rate:             utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.rate }"),
			BitRate:          utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.bit_rate }"),
			ExplicitLanguage: utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.explicit_language }"),
			ContextLanguage:  utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.context_language }"),
			LoudnessRatio:    utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.loudness_ratio }"),
		},
		Request: SpeechRequestOptionsRequest{
			RequestID:             requestID,
			Text:                  opts.Input,
			TextType:              utils.GetByJSONPath[*string](opts.ExtraBody, "{ .request.text_type }"),
			SilenceDuration:       utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .request.silence_duration }"),
			WithTimestamp:         utils.GetByJSONPath[*string](opts.ExtraBody, "{ .request.with_timestamp }"),
			Operation:             operation,
			ExtraParam:            utils.GetByJSONPath[*string](opts.ExtraBody, "{ .request.extra_param }"),
			DisableMarkdownFilter: utils.GetByJSONPath[*bool](opts.ExtraBody, "{ .request.disable_markdown_filter }"),
			EnableLatexTone:       utils.GetByJSONPath[*bool](opts.ExtraBody, "{ .request.enable_latex_tn }"),
			CacheConfig:           utils.GetByJSONPath[map[string]any](opts.ExtraBody, "{ .request.cache_config }"),
		},
	}

	jsonBytes, err := json.Marshal(newReqParams)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, "https://openspeech.bytedance.com/api/v1/tts", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	req.Header.Set("Authorization", "Bearer;"+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		switch {
		case strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json"):
			return mo.Err[any](apierrors.
				NewUpstreamError(resp.StatusCode).
				WithDetail(utils.NewJSONResponseError(resp.StatusCode, resp.Body).OrEmpty().Error()))
		case strings.HasPrefix(resp.Header.Get("Content-Type"), "text/"):
			return mo.Err[any](apierrors.
				NewUpstreamError(resp.StatusCode).
				WithDetail(utils.NewTextResponseError(resp.StatusCode, resp.Body).OrEmpty().Error()))
		default:
			slog.Warn("unknown upstream error with unknown Content-Type",
				slog.Int("status", resp.StatusCode),
				slog.String("content_type", resp.Header.Get("Content-Type")),
				slog.String("content_length", resp.Header.Get("Content-Length")),
			)
		}
	}

	var resBody map[string]any

	err = json.NewDecoder(resp.Body).Decode(&resBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	audioBase64String := utils.GetByJSONPath[string](resBody, "{ .data }")
	if audioBase64String == "" {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("upstream returned empty audio base64 string").WithCaller())
	}

	audioBytes, err := base64.StdEncoding.DecodeString(audioBase64String)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", audioBytes))
}
