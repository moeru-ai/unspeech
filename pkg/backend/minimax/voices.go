package minimax

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
	"github.com/samber/mo"
)

// GetVoiceReq Request for getting voice list
type GetVoiceReq struct {
	VoiceType string `json:"voice_type"`
}

// SystemVoice System voice
type SystemVoice struct {
	VoiceID     string   `json:"voice_id"`
	VoiceName   string   `json:"voice_name"`
	Description []string `json:"description"`
}

// VoiceCloning Voice cloning
type VoiceCloning struct {
	VoiceID     string   `json:"voice_id"`
	Description []string `json:"description"`
	CreatedTime string   `json:"created_time"`
}

// VoiceGeneration Voice generation
type VoiceGeneration struct {
	VoiceID     string   `json:"voice_id"`
	Description []string `json:"description"`
	CreatedTime string   `json:"created_time"`
}

// BaseResp Base response
type BaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// GetVoiceResp Response for getting voice list
type GetVoiceResp struct {
	SystemVoice      []SystemVoice     `json:"system_voice"`
	VoiceCloning    []VoiceCloning   `json:"voice_cloning"`
	VoiceGeneration []VoiceGeneration `json:"voice_generation"`
	BaseResp        BaseResp          `json:"base_resp"`
}

var (
	// Supported audio formats
	formats = []types.VoiceFormat{
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg"},
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm"},
		{Name: "FLAC", Extension: ".flac", MimeType: "audio/flac"},
		{Name: "WAV", Extension: ".wav", MimeType: "audio/wav"},
	}
)

// HandleVoices handles getting voice list requests
func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	// Get token
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	// Build request
	reqBody := GetVoiceReq{
		VoiceType: "all",
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// Create request
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.minimaxi.com/v1/get_voice",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// Set request headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	defer func() { _ = resp.Body.Close() }()

	// Check HTTP status code
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		switch {
		case strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json"):
			return mo.Err[any](apierrors.
				NewUpstreamError(resp.StatusCode).
				WithDetail(utils.NewJSONResponseError(resp.StatusCode, resp.Body).OrEmpty().Error()))
		default:
			slog.Warn("unknown upstream error with unknown Content-Type",
				slog.Int("status", resp.StatusCode),
				slog.String("content_type", resp.Header.Get("Content-Type")),
			)
		}
	}

	// Parse response
	var voiceResp GetVoiceResp
	err = json.NewDecoder(resp.Body).Decode(&voiceResp)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	// Check business status code
	if voiceResp.BaseResp.StatusCode != 0 {
		return mo.Err[any](handleMinimaxError(voiceResp.BaseResp.StatusCode, voiceResp.BaseResp.StatusMsg))
	}

	// Convert voice list
	voices := make([]types.Voice, 0, len(voiceResp.SystemVoice)+len(voiceResp.VoiceCloning)+len(voiceResp.VoiceGeneration))

	// Add system voices
	for _, v := range voiceResp.SystemVoice {
		voices = append(voices, types.Voice{
			ID:          v.VoiceID,
			Name:        v.VoiceName,
			Description: strings.Join(v.Description, ", "),
			Labels: map[string]any{
				"type": "system",
			},
			Tags:              []string{"system"},
			Languages:         []types.VoiceLanguage{{Title: "Auto", Code: "auto"}},
			Formats:           formats,
			CompatibleModels:  []string{"speech-2.8-turbo", "speech-2.8-hd", "speech-2.6-turbo", "speech-2.6-hd"},
			PredefinedOptions: map[string]any{},
		})
	}

	// Add voice cloning voices
	for _, v := range voiceResp.VoiceCloning {
		voices = append(voices, types.Voice{
			ID:          v.VoiceID,
			Name:        v.VoiceID,
			Description: strings.Join(v.Description, ", "),
			Labels: map[string]any{
				"type":        "voice_cloning",
				"createdTime": v.CreatedTime,
			},
			Tags:              []string{"voice_cloning"},
			Languages:         []types.VoiceLanguage{{Title: "Auto", Code: "auto"}},
			Formats:           formats,
			CompatibleModels:  []string{"speech-2.8-turbo", "speech-2.8-hd"},
			PredefinedOptions: map[string]any{},
		})
	}

	// Add voice generation voices
	for _, v := range voiceResp.VoiceGeneration {
		voices = append(voices, types.Voice{
			ID:          v.VoiceID,
			Name:        v.VoiceID,
			Description: strings.Join(v.Description, ", "),
			Labels: map[string]any{
				"type":        "voice_generation",
				"createdTime": v.CreatedTime,
			},
			Tags:              []string{"voice_generation"},
			Languages:         []types.VoiceLanguage{{Title: "Auto", Code: "auto"}},
			Formats:           formats,
			CompatibleModels:  []string{"speech-2.8-turbo", "speech-2.8-hd"},
			PredefinedOptions: map[string]any{},
		})
	}

	return mo.Ok[any](types.ListVoicesResponse{
		Voices: voices,
	})
}
