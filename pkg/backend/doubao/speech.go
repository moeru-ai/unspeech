package doubao

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

// DoubaoSpeechRequest represents the request structure for Doubao TTS API v3
// Reference: https://www.volcengine.com/docs/6561/1329505
// Note: Although the protocol is v3, the endpoint URL remains /api/v1/tts
type DoubaoSpeechRequest struct {
	App *DoubaoSpeechRequestApp `json:"app,omitempty"`
	// User configuration
	User *DoubaoSpeechRequestUser `json:"user,omitempty"`
	// Audio synthesis parameters
	Audio *DoubaoSpeechRequestAudio `json:"audio,omitempty"`
	// Request metadata
	Request *DoubaoSpeechRequestRequest `json:"request,omitempty"`
}

// DoubaoSpeechRequestApp contains application credentials
type DoubaoSpeechRequestApp struct {
	AppID   string `json:"appid,omitempty"`
	Token   string `json:"token,omitempty"`
	Cluster string `json:"cluster,omitempty"`
}

// DoubaoSpeechRequestUser contains user information
type DoubaoSpeechRequestUser struct {
	UserID string `json:"uid,omitempty"`
}

// DoubaoSpeechRequestAudio contains audio synthesis parameters
type DoubaoSpeechRequestAudio struct {
	VoiceType        string   `json:"voice_type,omitempty"`
	Encoding         *string  `json:"encoding,omitempty"`
	SpeedRatio       *float64 `json:"speed_ratio,omitempty"`
	Rate             *int     `json:"rate,omitempty"`
	BitRate          *int     `json:"bit_rate,omitempty"`
	ExplicitLanguage *string  `json:"explicit_language,omitempty"`
	ContextLanguage  *string  `json:"context_language,omitempty"`
	LoudnessRatio    *float64 `json:"loudness_ratio,omitempty"`
	Pitch            *float64 `json:"pitch,omitempty"`
	Emotion          *string  `json:"emotion,omitempty"`
	ResourceID       *string  `json:"resource_id,omitempty"`
}

// DoubaoSpeechRequestRequest contains request metadata
type DoubaoSpeechRequestRequest struct {
	RequestID             string         `json:"reqid,omitempty"`
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

// HandleSpeech handles the speech synthesis request for Doubao TTS
func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opts := options.MustGet()

	// Extract token from Authorization header
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	// Doubao TTS uses doubao_tts cluster by default
	cluster := utils.GetByJSONPath[string](opts.ExtraBody, "{ .app.cluster }")
	if cluster == "" {
		cluster = "doubao_tts"
	}

	// Generate or get user ID
	userID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .user.uid }")
	if userID == "" {
		userID = uuid.New().String()
	}

	// Generate or get request ID
	requestID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .request.reqid }")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// Get operation type, default to "query"
	operation := utils.GetByJSONPath[*string](opts.ExtraBody, "{ .request.operation }")
	if operation == nil || *operation == "" {
		operation = lo.ToPtr("query")
	}

	// Get speed ratio, default to 1.0
	speedRatio := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.speed_ratio }")
	if speedRatio == nil || *speedRatio == 0 {
		speedRatio = lo.ToPtr(1.0)
	}

	// Build the request payload
	reqBody := &DoubaoSpeechRequest{
		App: &DoubaoSpeechRequestApp{
			AppID:   utils.GetByJSONPath[string](opts.ExtraBody, "{ .app.appid }"),
			Token:   token,
			Cluster: cluster,
		},
		User: &DoubaoSpeechRequestUser{
			UserID: userID,
		},
		Audio: &DoubaoSpeechRequestAudio{
			VoiceType:        opts.Voice,
			Encoding:         lo.Ternary(opts.ResponseFormat != "", lo.ToPtr(opts.ResponseFormat), lo.ToPtr("mp3")),
			SpeedRatio:       speedRatio,
			Rate:             utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.rate }"),
			BitRate:          utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.bit_rate }"),
			ExplicitLanguage: utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.explicit_language }"),
			ContextLanguage:  utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.context_language }"),
			LoudnessRatio:    utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.loudness_ratio }"),
			Pitch:            utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.pitch }"),
			Emotion:          utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.emotion }"),
			ResourceID:       utils.GetByJSONPath[*string](opts.ExtraBody, "{ .audio.resource_id }"),
		},
		Request: &DoubaoSpeechRequestRequest{
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

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, "https://openspeech.bytedance.com/api/v1/tts", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// Set authorization header - Doubao uses Bearer;token format
	req.Header.Set("Authorization", "Bearer;"+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
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

	// Extract base64 encoded audio data from response
	audioBase64String := utils.GetByJSONPath[string](resBody, "{ .data }")
	if audioBase64String == "" {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("upstream returned empty audio base64 string").WithCaller())
	}

	// Decode base64 audio to binary
	audioBytes, err := base64.StdEncoding.DecodeString(audioBase64String)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", audioBytes))
}
