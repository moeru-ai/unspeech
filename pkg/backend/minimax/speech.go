package minimax

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/mo"
)

// VoiceSetting MiniMax voice settings
type VoiceSetting struct {
	VoiceID           string  `json:"voice_id"`
	Speed             float64 `json:"speed,omitempty"`
	Vol               float64 `json:"vol,omitempty"`
	Pitch             float64 `json:"pitch,omitempty"`
	Emotion           string  `json:"emotion,omitempty"`
	TextNormalization string  `json:"text_normalization,omitempty"`
	LatexRead         string  `json:"latex_read,omitempty"`
}

// AudioSetting MiniMax audio settings
type AudioSetting struct {
	SampleRate int    `json:"sample_rate,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"`
	Format     string `json:"format,omitempty"`
	Channel    int    `json:"channel,omitempty"`
}

// TTSRequest MiniMax TTS request
type TTSRequest struct {
	Model        string        `json:"model"`
	Text         string        `json:"text"`
	Stream       bool          `json:"stream,omitempty"`
	VoiceSetting *VoiceSetting `json:"voice_setting,omitempty"`
	AudioSetting *AudioSetting `json:"audio_setting,omitempty"`
	OutputFormat string        `json:"output_format,omitempty"`
}

// TTSResponseData MiniMax TTS response data
type TTSResponseData struct {
	Audio        string `json:"audio"`
	SubtitleFile string `json:"subtitle_file,omitempty"`
	Status       int    `json:"status"`
}

// TTSResponseExtraInfo MiniMax TTS response extra info
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

// TTSResponseBaseResp MiniMax TTS response base info
type TTSResponseBaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// TTSResponse MiniMax TTS response
type TTSResponse struct {
	Data      TTSResponseData      `json:"data"`
	TraceID   string               `json:"trace_id"`
	ExtraInfo TTSResponseExtraInfo `json:"extra_info"`
	BaseResp  TTSResponseBaseResp  `json:"base_resp"`
}

// HandleSpeech handles MiniMax TTS requests
func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	opts := options.MustGet()

	// Get token
	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	// Get stream parameter from ExtraBody
	stream := utils.GetByJSONPath[bool](opts.ExtraBody, "{ .stream }")

	// If streaming, use streaming handler
	if stream {
		return handleStreamingSpeech(c, token, opts)
	}

	reqBody := buildTTSRequest(opts, false)

	// Serialize request body
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// Create request
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.minimaxi.com/v1/t2a_v2",
		bytes.NewBuffer(jsonBytes),
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
		return handleHTTPError(resp)
	}

	// Parse response
	var ttsResp TTSResponse
	err = json.NewDecoder(resp.Body).Decode(&ttsResp)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	// Check business status code
	if ttsResp.BaseResp.StatusCode != 0 {
		return mo.Err[any](handleMinimaxError(ttsResp.BaseResp.StatusCode, ttsResp.BaseResp.StatusMsg))
	}

	// Decode hex audio
	audioBytes, err := hex.DecodeString(ttsResp.Data.Audio)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("failed to decode hex audio: " + err.Error()).WithError(err).WithCaller())
	}

	// Determine content type
	contentType := getContentType(reqBody.AudioSetting)

	return mo.Ok[any](c.Blob(http.StatusOK, contentType, audioBytes))
}

func buildTTSRequest(opts types.SpeechRequestOptions, stream bool) TTSRequest {
	reqBody := TTSRequest{
		Model:        opts.Model,
		Text:         opts.Input,
		Stream:       stream,
		OutputFormat: "hex",
	}

	if opts.Voice != "" || opts.Speed != 0 {
		reqBody.VoiceSetting = &VoiceSetting{
			VoiceID: opts.Voice,
			Speed:   opts.Speed,
		}
	}
	if opts.ResponseFormat != "" {
		reqBody.AudioSetting = &AudioSetting{Format: opts.ResponseFormat}
	}

	// Provider-specific extra_body values take precedence over OpenAI fields.
	buildVoiceSettings(opts.ExtraBody, &reqBody)
	buildAudioSettings(opts.ExtraBody, &reqBody)

	return reqBody
}

// buildVoiceSettings builds VoiceSetting from ExtraBody
func buildVoiceSettings(extraBody map[string]any, reqBody *TTSRequest) {
	if speed := utils.GetByJSONPath[*float64](extraBody, "{ .speed }"); speed != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Speed = *speed
	}

	if vol := utils.GetByJSONPath[*float64](extraBody, "{ .vol }"); vol != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Vol = *vol
	}

	if pitch := utils.GetByJSONPath[*float64](extraBody, "{ .pitch }"); pitch != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Pitch = *pitch
	}

	if emotion := utils.GetByJSONPath[*string](extraBody, "{ .emotion }"); emotion != nil {
		if reqBody.VoiceSetting == nil {
			reqBody.VoiceSetting = &VoiceSetting{}
		}
		reqBody.VoiceSetting.Emotion = *emotion
	}
}

// buildAudioSettings builds AudioSetting from ExtraBody
func buildAudioSettings(extraBody map[string]any, reqBody *TTSRequest) {
	if sampleRate := utils.GetByJSONPath[*int](extraBody, "{ .sample_rate }"); sampleRate != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.SampleRate = *sampleRate
	}

	if bitrate := utils.GetByJSONPath[*int](extraBody, "{ .bitrate }"); bitrate != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Bitrate = *bitrate
	}

	if format := utils.GetByJSONPath[*string](extraBody, "{ .format }"); format != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Format = *format
	}

	if channel := utils.GetByJSONPath[*int](extraBody, "{ .channel }"); channel != nil {
		if reqBody.AudioSetting == nil {
			reqBody.AudioSetting = &AudioSetting{}
		}
		reqBody.AudioSetting.Channel = *channel
	}
}

// getContentType returns MIME type based on audio format
func getContentType(audioSetting *AudioSetting) string {
	if audioSetting == nil || audioSetting.Format == "" {
		return "audio/mp3"
	}

	contentTypes := map[string]string{
		"pcm":  "audio/pcm",
		"wav":  "audio/wav",
		"flac": "audio/flac",
		"mp3":  "audio/mpeg",
	}

	if ct, ok := contentTypes[audioSetting.Format]; ok {
		return ct
	}
	return "audio/mpeg"
}

// handleHTTPError handles HTTP errors from upstream
func handleHTTPError(resp *http.Response) mo.Result[any] {
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
	return mo.Err[any](apierrors.NewUpstreamError(resp.StatusCode))
}

// handleMinimaxError handles MiniMax error codes
func handleMinimaxError(code int, msg string) *apierrors.Error {
	var httpStatus int
	switch code {
	case 1004: // Auth failed
		httpStatus = http.StatusUnauthorized
	case 1002, 1039: // Rate limit
		httpStatus = http.StatusTooManyRequests
	case 1042, 2013: // Invalid parameter
		httpStatus = http.StatusBadRequest
	case 1001: // Timeout
		httpStatus = http.StatusGatewayTimeout
	default:
		httpStatus = http.StatusBadGateway
	}
	return apierrors.NewUpstreamError(httpStatus).WithDetailf("minimax error: %d - %s", code, msg)
}

// handleStreamingSpeech handles streaming TTS requests
func handleStreamingSpeech(c echo.Context, token string, opts types.SpeechRequestOptions) mo.Result[any] {
	reqBody := buildTTSRequest(opts, true)

	// Serialize request body
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	// Create request
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.minimaxi.com/v1/t2a_v2",
		bytes.NewBuffer(jsonBytes),
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
		return handleHTTPError(resp)
	}

	// Streaming: read response until status == 2
	decoder := json.NewDecoder(resp.Body)
	c.Response().Header().Set(echo.HeaderContentType, getContentType(reqBody.AudioSetting))

	for {
		var ttsResp TTSResponse
		if err := decoder.Decode(&ttsResp); err != nil {
			if err == io.EOF {
				return handleStreamError(c, apierrors.NewErrBadGateway().WithDetail("upstream stream ended before completion").WithCaller())
			}
			return handleStreamError(c, apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
		}

		// Check business status code
		if ttsResp.BaseResp.StatusCode != 0 {
			return handleStreamError(c, handleMinimaxError(ttsResp.BaseResp.StatusCode, ttsResp.BaseResp.StatusMsg))
		}

		if ttsResp.Data.Audio != "" {
			audioBytes, decodeErr := hex.DecodeString(ttsResp.Data.Audio)
			if decodeErr != nil {
				return handleStreamError(c, apierrors.NewErrInternal().WithDetail("failed to decode hex audio: "+decodeErr.Error()).WithError(decodeErr).WithCaller())
			}

			if _, writeErr := c.Response().Write(audioBytes); writeErr != nil {
				return handleStreamError(c, apierrors.NewErrInternal().WithDetail(writeErr.Error()).WithError(writeErr).WithCaller())
			}
			c.Response().Flush()
		}

		// status == 2 means synthesis complete
		if ttsResp.Data.Status == 2 {
			break
		}
	}

	return mo.Ok[any](nil)
}

func handleStreamError(c echo.Context, err *apierrors.Error) mo.Result[any] {
	if !c.Response().Committed {
		return mo.Err[any](err)
	}

	// Once audio bytes have been sent, an HTTP error response would corrupt the
	// stream. Abort the connection so Fetch/HTTP clients fail while reading the
	// body instead of accepting truncated audio as a successful 200 response.
	panic(http.ErrAbortHandler)
}
