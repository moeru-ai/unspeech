package minimax

import (
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

// GetVoiceReq 获取音色列表请求
type GetVoiceReq struct {
	VoiceType string `json:"voice_type"`
}

// SystemVoice 系统音色
type SystemVoice struct {
	VoiceID     string   `json:"voice_id"`
	VoiceName   string   `json:"voice_name"`
	Description []string `json:"description"`
}

// VoiceCloning 快速复刻音色
type VoiceCloning struct {
	VoiceID     string   `json:"voice_id"`
	Description []string `json:"description"`
	CreatedTime string   `json:"created_time"`
}

// VoiceGeneration 文生音色
type VoiceGeneration struct {
	VoiceID     string   `json:"voice_id"`
	Description []string `json:"description"`
	CreatedTime string   `json:"created_time"`
}

// BaseResp 基础响应
type BaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// GetVoiceResp 获取音色列表响应
type GetVoiceResp struct {
	SystemVoice      []SystemVoice     `json:"system_voice"`
	VoiceCloning    []VoiceCloning   `json:"voice_cloning"`
	VoiceGeneration []VoiceGeneration `json:"voice_generation"`
	BaseResp        BaseResp          `json:"base_resp"`
}

var (
	// 支持的音频格式
	formats = []types.VoiceFormat{
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg"},
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm"},
		{Name: "FLAC", Extension: ".flac", MimeType: "audio/flac"},
		{Name: "WAV", Extension: ".wav", MimeType: "audio/wav"},
	}
)

// HandleVoices 处理获取音色列表请求
func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	// 获取 token
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	// 构建请求
	reqBody := GetVoiceReq{
		VoiceType: "all",
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// 创建请求
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.minimaxi.com/v1/get_voice",
		strings.NewReader(string(jsonBytes)),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	defer func() { _ = resp.Body.Close() }()

	// 检查 HTTP 状态码
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

	// 解析响应
	var voiceResp GetVoiceResp
	err = json.NewDecoder(resp.Body).Decode(&voiceResp)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	// 检查业务状态码
	if voiceResp.BaseResp.StatusCode != 0 {
		return mo.Err[any](handleMinimaxError(voiceResp.BaseResp.StatusCode, voiceResp.BaseResp.StatusMsg))
	}

	// 转换音色列表
	voices := make([]types.Voice, 0, len(voiceResp.SystemVoice)+len(voiceResp.VoiceCloning)+len(voiceResp.VoiceGeneration))

	// 添加系统音色
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

	// 添加快速复刻音色
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

	// 添加文生音色
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
