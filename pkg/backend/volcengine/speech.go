package volcengine

import (
	"bufio"
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

	appID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .app.appid }")

	if appID == "" {
		return handleSpeechV3(c, opts)
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	resourceID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .resource_id }")
	if resourceID == "" {
		if voices, err := ListVoices(c.Request().Context(), ""); err == nil {
			for _, v := range voices {
				if v.ID == opts.Voice {
					if len(v.CompatibleModels) > 0 {
						resourceID = v.CompatibleModels[0]
					}
					break
				}
			}
		}
	}
	if resourceID == "" {
		resourceID = "seed-tts-2.0"
	}

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
			AppID:   appID,
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

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, "https://openspeech.bytedance.com/api/v3/tts/unidirectional", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer;"+token)
	req.Header.Set("X-Api-App-Id", appID)
	req.Header.Set("X-Api-Access-Key", token)
	req.Header.Set("X-Api-Resource-Id", resourceID)
	req.Header.Set("X-Api-Request-Id", requestID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	defer func() { _ = resp.Body.Close() }()

	slog.Info("volcengine v3 request",
		slog.String("endpoint", "https://openspeech.bytedance.com/api/v3/tts/unidirectional"),
		slog.String("resource_id", resourceID),
		slog.String("voice_type", opts.Voice),
		slog.Int("status", resp.StatusCode),
		slog.String("logid", resp.Header.Get("X-Tt-Logid")),
	)

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

	var audioBytes []byte
	scanner := bufio.NewScanner(resp.Body)
	lineCount := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		lineCount++
		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			slog.Warn("volcengine v3: skip malformed line",
				slog.Int("line_no", lineCount),
				slog.String("line", string(line[:min(len(line), 200)])))
			continue
		}
		// Debug: log raw keys of first 3 lines
		if lineCount <= 3 {
			keys := make([]string, 0, len(raw))
			for k := range raw {
				keys = append(keys, k)
			}
			slog.Info("volcengine v3: debug line keys",
				slog.Int("line_no", lineCount),
				slog.Any("keys", keys),
				slog.String("raw", string(line[:min(len(line), 300)])))
		}
		b64 := ""
		if data, ok := raw["data"].(string); ok {
			b64 = data
		} else if msg, ok := raw["payload_msg"].(map[string]any); ok {
			if data, ok := msg["data"].(string); ok {
				b64 = data
			}
		}
		if b64 == "" {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			slog.Warn("volcengine v3: skip malformed base64 chunk", slog.String("err", err.Error()))
			continue
		}
		audioBytes = append(audioBytes, decoded...)
	}

	if err := scanner.Err(); err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("volcengine v3: read streaming body: " + err.Error()).WithCaller())
	}

	if len(audioBytes) == 0 {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("upstream returned empty audio").WithCaller())
	}

	return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", audioBytes))
}
