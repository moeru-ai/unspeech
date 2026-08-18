package stepfun

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/samber/mo"
)

const (
	modelStepAudio25TTS = "stepaudio-2.5-tts"
	modelStepTTS2       = "step-tts-2"
	modelStepTTSMini    = "step-tts-mini"
	defaultSampleRate   = 24000
)

var (
	languages = []types.VoiceLanguage{
		{Code: "zh-CN", Title: "Chinese"},
		{Code: "en", Title: "English"},
		{Code: "ja", Title: "Japanese"},
	}

	formats = []types.VoiceFormat{
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: defaultSampleRate, FormatCode: "mp3"},
		{Name: "WAV", Extension: ".wav", MimeType: "audio/wav", SampleRate: defaultSampleRate, FormatCode: "wav"},
		{Name: "FLAC", Extension: ".flac", MimeType: "audio/flac", SampleRate: defaultSampleRate, FormatCode: "flac"},
		{Name: "Opus", Extension: ".opus", MimeType: "audio/opus", SampleRate: defaultSampleRate, FormatCode: "opus"},
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/L16", SampleRate: defaultSampleRate, FormatCode: "pcm"},
	}

	voiceIDs = []struct {
		ID     string
		Name   string
		Models []string
		Scene  string
	}{
		{"vibrant-youth", "Vibrant Youth", []string{modelStepAudio25TTS, modelStepTTS2}, "有声书、视频配音"},
		{"lively-girl", "Lively Girl", []string{modelStepAudio25TTS, modelStepTTS2}, "有声书、视频配音"},
		{"soft-spoken-gentleman", "Soft-spoken Gentleman", []string{modelStepAudio25TTS, modelStepTTS2}, "情感陪伴、有声书"},
		{"magnetic-voiced-male", "Magnetic-voiced Male", []string{modelStepAudio25TTS, modelStepTTS2}, "有声书、视频配音"},
		{"zixinnansheng", "自信男声", []string{modelStepAudio25TTS, modelStepTTS2}, "有声书、情感陪伴、教育与培训、营销"},
		{"elegantgentle-female", "气质温婉", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "客服与业务办理、口播（解说、新闻）、教育与培训、情感陪伴"},
		{"livelybreezy-female", "活力轻快", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "情感陪伴、客服与业务办理、教育与培训、营销"},
		{"wenrounansheng", "温柔男声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "口播（解说、新闻）、情感陪伴、客服与业务办理、教育与培训"},
		{"wenrougongzi", "温柔公子", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "情感陪伴、有声书"},
		{"yuanqinansheng", "元气男声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "有声书、口播（解说、新闻）、客服与业务办理"},
		{"jingdiannvsheng", "经典女声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "客服与业务办理、情感陪伴"},
		{"wenroushunv", "温柔熟女", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "客服与业务办理、口播（解说、新闻）、教育与培训"},
		{"tianmeinvsheng", "甜美女声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "情感陪伴、客服与业务办理"},
		{"qingchunshaonv", "清纯少女", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "客服与业务办理、语音助手"},
		{"cixingnansheng", "磁性男声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "有声书、情感陪伴"},
		{"yuanqishaonv", "元气少女", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "有声书、情感陪伴、语音助手"},
		{"linjiajiejie", "邻家姐姐", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "口播（解说、新闻）、情感陪伴、语音助手、视频配音"},
		{"zhengpaiqingnian", "正派青年", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "营销、有声书"},
		{"qingniandaxuesheng", "青年大学生", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "口播（解说、新闻）"},
		{"boyinnansheng", "播音男声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "有声书、口播（解说、新闻）"},
		{"ruyananshi", "儒雅男士", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "有声书、情感陪伴、口播（解说、新闻）、语音助手"},
		{"shenchennanyin", "深沉男音", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "情感陪伴、有声书"},
		{"qinqienvsheng", "亲切女声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "口播（解说、新闻）"},
		{"wenrounvsheng", "温柔女声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "有声书、情感陪伴"},
		{"jilingshaonv", "机灵少女", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "语音助手、口播（解说、新闻）"},
		{"ruanmengnvsheng", "软萌女声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "情感陪伴、语音助手、视频配音"},
		{"youyanvsheng", "优雅女声", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "视频配音"},
		{"lengyanyujie", "冷艳御姐", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "视频配音"},
		{"shuangkuaijiejie", "爽快姐姐", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "口播（解说、新闻）"},
		{"wenjingxuejie", "文静学姐", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "口播（解说、新闻）"},
		{"linjiameimei", "邻家妹妹", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "视频配音、口播（解说、新闻）、语音助手"},
		{"zhixingjiejie", "知性姐姐", []string{modelStepAudio25TTS, modelStepTTS2, modelStepTTSMini}, "视频配音、口播（解说、新闻）、语音助手"},
		{"shuangkuainansheng", "爽快男声", []string{modelStepAudio25TTS, modelStepTTS2}, "客服与业务办理、语音助手"},
		{"ganliannvsheng", "干练女声", []string{modelStepAudio25TTS, modelStepTTS2}, "客服与业务办理、语音助手"},
		{"qinhenvsheng", "亲和女声", []string{modelStepAudio25TTS, modelStepTTS2}, "客服与业务办理、语音助手"},
		{"huolinvsheng", "活力女声", []string{modelStepAudio25TTS, modelStepTTS2}, "客服与业务办理、语音助手"},
	}
)

func ListVoices(_ context.Context) ([]types.Voice, error) {
	voices := make([]types.Voice, len(voiceIDs))
	for i, v := range voiceIDs {
		voices[i] = types.Voice{
			ID:               v.ID,
			Name:             v.Name,
			Description:      v.Scene,
			Labels:           map[string]any{"provider": "stepfun"},
			Tags:             splitSceneTags(v.Scene),
			Languages:        languages,
			Formats:          formats,
			CompatibleModels: v.Models,
		}
	}

	return voices, nil
}

func HandleVoices(c echo.Context, _ mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	voices, err := ListVoices(c.Request().Context())
	if err != nil {
		return mo.Err[any](err)
	}

	return mo.Ok[any](types.ListVoicesResponse{Voices: voices})
}

func splitSceneTags(scene string) []string {
	tags := make([]string, 0)
	start := 0

	for i, r := range scene {
		if r == '、' {
			if start < i {
				tags = append(tags, scene[start:i])
			}

			start = i + len(string(r))
		}
	}
	if start < len(scene) {
		tags = append(tags, scene[start:])
	}

	return tags
}
