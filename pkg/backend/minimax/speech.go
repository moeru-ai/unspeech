package minimax

import (
	"bytes"
	"encoding/hex"
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

// VoiceSetting MiniMax 音色设置
type VoiceSetting struct {
	VoiceID           string  `json:"voice_id"`
	Speed             float64 `json:"speed,omitempty"`
	Vol               float64 `json:"vol,omitempty"`
	Pitch             float64 `json:"pitch,omitempty"`
	Emotion           string  `json:"emotion,omitempty"`
	TextNormalization string  `json:"text_normalization,omitempty"`
	LatexRead         string  `json:"latex_read,omitempty"`
}

// AudioSetting MiniMax 音频设置
type AudioSetting struct {
	SampleRate int    `json:"sample_rate,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"`
	Format     string `json:"format,omitempty"`
	Channel    int    `json:"channel,omitempty"`
}

// TTSRequest MiniMax TTS 请求
type TTSRequest struct {
	Model        string        `json:"model"`
	Text         string        `json:"text"`
	Stream       bool          `json:"stream,omitempty"`
	VoiceSetting *VoiceSetting `json:"voice_setting,omitempty"`
	AudioSetting *AudioSetting `json:"audio_setting,omitempty"`
	OutputFormat string        `json:"output_format,omitempty"`
}

// TTSResponseData MiniMax TTS 响应数据
type TTSResponseData struct {
	Audio        string `json:"audio"`
	SubtitleFile string `json:"subtitle_file,omitempty"`
	Status       int    `json:"status"`
}

// TTSResponseExtraInfo MiniMax TTS 响应额外信息
type TTSResponseExtraInfo struct {
	AudioLength     int    `json:"audio_length"`
	AudioSampleRate int    `json:"audio_sample_rate"`
	AudioSize       int    `json:"audio_size"`
	Bitrate         int    `json:"bitrate"`
	AudioFormat     string `json:"audio_format"`
	AudioChannel    int    `json:"audio_channel"`
	WordCount       int    `json:"word_count"`
	UsageCharacters int    `json:"usage_characters"`
}

// TTSResponseBaseResp MiniMax TTS 响应基础信息
type TTSResponseBaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// TTSResponse MiniMax TTS 响应
type TTSResponse struct {
	Data      TTSResponseData      `json:"data"`
	TraceID   string              `json:"trace_id"`
	ExtraInfo TTSResponseExtraInfo `json:"extra_info"`
	BaseResp  TTSResponseBaseResp  `json:"base_resp"`
}

// HandleSpeech 处理 MiniMax TTS 请求
func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opts := options.MustGet()

	// 获取 token
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	// 从 ExtraBody 获取 stream 参数
	stream := utils.GetByJSONPath[bool](opts.ExtraBody, "{ .stream }")

	// 如果是流式请求，使用流式处理
	if stream {
		return handleStreamingSpeech(c, token, opts)
	}

	// 构建 MiniMax 请求
	reqBody := TTSRequest{
		Model:        opts.Model,
		Text:         opts.Input,
		Stream:       false,
		OutputFormat: "hex",
	}

	// 设置 voice_id（从用户传入的 voice 字段）
	voiceID := opts.Voice
	if voiceID != "" {
		reqBody.VoiceSetting = &VoiceSetting{
			VoiceID: voiceID,
		}
	}

	// 从 ExtraBody 获取其他参数
	if speed := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .speed }"); speed != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Speed = *speed
	}

	if vol := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .vol }"); vol != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Vol = *vol
	}

	if pitch := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .pitch }"); pitch != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Pitch = *pitch
	}

	if emotion := utils.GetByJSONPath[*string](opts.ExtraBody, "{ .emotion }"); emotion != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Emotion = *emotion
	}

	// 音频设置
	if sampleRate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .sample_rate }"); sampleRate != nil {
		reqBody.AudioSetting = &AudioSetting{}
		reqBody.AudioSetting.SampleRate = *sampleRate
	}

	if bitrate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .bitrate }"); bitrate != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Bitrate = *bitrate
	}

	if format := utils.GetByJSONPath[*string](opts.ExtraBody, "{ .format }"); format != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Format = *format
	}

	// 序列化请求体
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// 创建请求
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.minimaxi.com/v1/t2a_v2",
		bytes.NewBuffer(jsonBytes),
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
		case strings.HasPrefix(resp.Header.Get("Content-Type"), "text/"):
			return mo.Err[any](apierrors.
				NewUpstreamError(resp.StatusCode).
				WithDetail(utils.NewTextResponseError(resp.StatusCode, resp.Body).OrEmpty().Error()))
		default:
			slog.Warn("unknown upstream error with unknown Content-Type",
				slog.Int("status", resp.StatusCode),
				slog.String("content_type", resp.Header.Get("Content-Type")),
			)
		}
	}

	// 解析响应
	var ttsResp TTSResponse
	err = json.NewDecoder(resp.Body).Decode(&ttsResp)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	// 检查业务状态码
	if ttsResp.BaseResp.StatusCode != 0 {
		return mo.Err[any](handleMinimaxError(ttsResp.BaseResp.StatusCode, ttsResp.BaseResp.StatusMsg))
	}

	// 解码 hex 音频
	audioBytes, err := hex.DecodeString(ttsResp.Data.Audio)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("failed to decode hex audio: "+err.Error()).WithError(err).WithCaller())
	}

	// 确定 Content-Type
	contentType := "audio/mp3"
	if reqBody.AudioSetting != nil && reqBody.AudioSetting.Format != "" {
		contentType = lo.Ternary(reqBody.AudioSetting.Format == "pcm", "audio/pcm",
			lo.Ternary(reqBody.AudioSetting.Format == "wav", "audio/wav",
				lo.Ternary(reqBody.AudioSetting.Format == "flac", "audio/flac", "audio/mp3")))
	}

	return mo.Ok[any](c.Blob(http.StatusOK, contentType, audioBytes))
}

// handleMinimaxError 处理 MiniMax 错误码
func handleMinimaxError(code int, msg string) *apierrors.Error {
	var httpStatus int
	switch code {
	case 1004: // 鉴权失败
		httpStatus = http.StatusUnauthorized
	case 1002, 1039: // 限流
		httpStatus = http.StatusTooManyRequests
	case 1042, 2013: // 参数错误
		httpStatus = http.StatusBadRequest
	case 1001: // 超时
		httpStatus = http.StatusGatewayTimeout
	default:
		httpStatus = http.StatusBadGateway
	}
	return apierrors.NewUpstreamError(httpStatus).WithDetailf("minimax error: %d - %s", code, msg)
}

// handleStreamingSpeech 处理流式 TTS 请求
func handleStreamingSpeech(c echo.Context, token string, opts types.SpeechRequestOptions) mo.Result[any] {
	// 构建 MiniMax 请求
	reqBody := TTSRequest{
		Model:        opts.Model,
		Text:         opts.Input,
		Stream:       true,
		OutputFormat: "hex",
	}

	// 设置 voice_id
	voiceID := opts.Voice
	if voiceID != "" {
		reqBody.VoiceSetting = &VoiceSetting{
			VoiceID: voiceID,
		}
	}

	// 从 ExtraBody 获取其他参数
	if speed := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .speed }"); speed != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Speed = *speed
	}

	if vol := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .vol }"); vol != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Vol = *vol
	}

	if pitch := utils.GetByJSONPath[*float64](opts.ExtraBody, "{ .pitch }"); pitch != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Pitch = *pitch
	}

	if emotion := utils.GetByJSONPath[*string](opts.ExtraBody, "{ .emotion }"); emotion != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Emotion = *emotion
	}

	// 音频设置
	if sampleRate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .sample_rate }"); sampleRate != nil {
		reqBody.AudioSetting = &AudioSetting{}
		reqBody.AudioSetting.SampleRate = *sampleRate
	}

	if bitrate := utils.GetByJSONPath[*int](opts.ExtraBody, "{ .bitrate }"); bitrate != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Bitrate = *bitrate
	}

	if format := utils.GetByJSONPath[*string](opts.ExtraBody, "{ .format }"); format != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Format = *format
	}

	// 序列化请求体
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// 创建请求
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.minimaxi.com/v1/t2a_v2",
		bytes.NewBuffer(jsonBytes),
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

	// 流式处理：持续读取响应直到 status == 2
	decoder := json.NewDecoder(resp.Body)
	audioHex := new(strings.Builder)

	for {
		var ttsResp TTSResponse
		if err := decoder.Decode(&ttsResp); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
		}

		// 检查业务状态码
		if ttsResp.BaseResp.StatusCode != 0 {
			return mo.Err[any](handleMinimaxError(ttsResp.BaseResp.StatusCode, ttsResp.BaseResp.StatusMsg))
		}

		// 追加音频数据
		audioHex.WriteString(ttsResp.Data.Audio)

		// status == 2 表示合成结束
		if ttsResp.Data.Status == 2 {
			break
		}
	}

	// 解码 hex 音频
	audioBytes, err := hex.DecodeString(audioHex.String())
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("failed to decode hex audio: "+err.Error()).WithError(err).WithCaller())
	}

	// 确定 Content-Type
	contentType := "audio/mp3"
	if reqBody.AudioSetting != nil && reqBody.AudioSetting.Format != "" {
		contentType = lo.Ternary(reqBody.AudioSetting.Format == "pcm", "audio/pcm",
			lo.Ternary(reqBody.AudioSetting.Format == "wav", "audio/wav",
				lo.Ternary(reqBody.AudioSetting.Format == "flac", "audio/flac", "audio/mp3")))
	}

	return mo.Ok[any](c.Blob(http.StatusOK, contentType, audioBytes))
}
