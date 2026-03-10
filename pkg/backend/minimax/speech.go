// Package minimax provides MiniMax TTS integration
package minimax

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
