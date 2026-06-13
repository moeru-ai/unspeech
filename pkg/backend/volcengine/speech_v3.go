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

type v3User struct {
	UID string `json:"uid"`
}

type v3Request struct {
	User      v3User      `json:"user"`
	Namespace string      `json:"namespace"`
	ReqParams v3ReqParams `json:"req_params"`
}

func handleSpeechV3(c echo.Context, opts types.SpeechRequestOptions) mo.Result[any] {
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	resourceID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .resource_id }")
	if resourceID == "" {
		resourceID = "seed-tts-2.0"
	}

	userID := utils.GetByJSONPath[string](opts.ExtraBody, "{ .user.uid }")
	if userID == "" {
		userID = uuid.New().String()
	}

	format := lo.Ternary(opts.ResponseFormat != "", opts.ResponseFormat, "mp3")

	sampleRate := utils.GetByJSONPath[int](opts.ExtraBody, "{ .audio.rate }")
	if sampleRate == 0 {
		sampleRate = 24000
	}

	audio := v3ReqParamsAudio{
		Format:     format,
		SampleRate: sampleRate,
	}

	if speechRate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.speech_rate }"); speechRate != nil {
		audio.SpeechRate = *speechRate
	}

	if loudnessRate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.loudness_rate }"); loudnessRate != nil {
		audio.LoudnessRate = *loudnessRate
	}

	if bitRate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .audio.bit_rate }"); bitRate != nil {
		audio.BitRate = *bitRate
	} else if format == "mp3" {
		audio.BitRate = 128000
	}

	if emotion := utils.GetByJSONPath[string](opts.ExtraBody, "{ .audio.emotion }"); emotion != "" {
		audio.Emotion = emotion
	}

	if emotionScale := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.emotion_scale }"); emotionScale != nil {
		audio.EmotionScale = *emotionScale
	} else {
		audio.EmotionScale = 4
	}

	params := v3ReqParams{
		Text:        opts.Input,
		Speaker:     opts.Voice,
		AudioParams: &audio,
	}

	if pitch := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .audio.pitch }"); pitch != nil {
		pitchJSON, _ := json.Marshal(map[string]any{
			"post_process": map[string]float64{
				"pitch": *pitch,
			},
		})
		params.Additions = string(pitchJSON)
	}

	v3Req := v3Request{
		User: v3User{
			UID: userID,
		},
		Namespace: "TTS",
		ReqParams: params,
	}

	jsonBytes, err := json.Marshal(v3Req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, "https://openspeech.bytedance.com/api/v3/tts/unidirectional", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", token)
	req.Header.Set("X-Api-Resource-Id", resourceID)

	slog.Info("volcengine v3 request",
		slog.String("endpoint", "https://openspeech.bytedance.com/api/v3/tts/unidirectional"),
		slog.String("resource_id", resourceID),
		slog.String("voice_type", opts.Voice),
		slog.String("auth_mode", "x-api-key"),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	defer func() { _ = resp.Body.Close() }()

	slog.Info("volcengine v3 response",
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

	var bodyBuf bytes.Buffer
	if _, err := bodyBuf.ReadFrom(resp.Body); err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("volcengine v3: read body: " + err.Error()).WithCaller())
	}
	bodyData := bodyBuf.Bytes()

	var audioBytes []byte
	scanner := bufio.NewScanner(bytes.NewReader(bodyData))
	lineCount := 0
	hasJSONLines := false
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
		hasJSONLines = true

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

		if code, ok := raw["code"]; ok {
			var codeNum int64
			switch v := code.(type) {
			case float64:
				codeNum = int64(v)
			case int64:
				codeNum = v
			case json.Number:
				codeNum, _ = v.Int64()
			}
			if codeNum != 0 && codeNum != 20000000 {
				msg := ""
				if m, ok := raw["message"].(string); ok {
					msg = m
				}
				return mo.Err[any](apierrors.NewUpstreamError(resp.StatusCode).WithDetail(msg))
			}
		}

		b64 := ""
		if data, ok := raw["data"].(string); ok {
			b64 = data
		} else if audio, ok := raw["audio"].(map[string]any); ok {
			if data, ok := audio["data"].(string); ok {
				b64 = data
			}
		}
		if b64 == "" {
			continue
		}

		b64 = strings.NewReplacer("\r", "", "\n", "").Replace(b64)

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

	if len(audioBytes) == 0 && !hasJSONLines {
		var raw map[string]any
		if err := json.Unmarshal(bodyData, &raw); err == nil {
			if code, ok := raw["code"]; ok {
				var codeNum int64
				switch v := code.(type) {
				case float64:
					codeNum = int64(v)
				case int64:
					codeNum = v
				case json.Number:
					codeNum, _ = v.Int64()
				}
				if codeNum != 0 && codeNum != 20000000 {
					msg := ""
					if m, ok := raw["message"].(string); ok {
						msg = m
					}
					return mo.Err[any](apierrors.NewUpstreamError(resp.StatusCode).WithDetail(msg))
				}
			}

			b64 := ""
			if data, ok := raw["data"].(string); ok {
				b64 = data
			} else if audio, ok := raw["audio"].(map[string]any); ok {
				if data, ok := audio["data"].(string); ok {
					b64 = data
				}
			}
			if b64 != "" {
				b64 = strings.NewReplacer("\r", "", "\n", "").Replace(b64)
				decoded, err := base64.StdEncoding.DecodeString(b64)
				if err == nil {
					audioBytes = append(audioBytes, decoded...)
				}
			}
		}
	}

	if len(audioBytes) == 0 {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("upstream returned empty audio").WithCaller())
	}

	return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", audioBytes))
}
