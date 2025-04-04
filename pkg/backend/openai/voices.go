package openai

import (
	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/samber/mo"
)

var (
	// Text to speech - OpenAI API
	// https://platform.openai.com/docs/guides/text-to-speech#supported-languages
	languages = []types.VoiceLanguage{
		{Code: "af-ZA", Title: "Afrikaans"},
		{Code: "ar-SA", Title: "Arabic"},
		{Code: "hy-AM", Title: "Armenian"},
		{Code: "az-AZ", Title: "Azerbaijani"},
		{Code: "be-BY", Title: "Belarusian"},
		{Code: "bs-BA", Title: "Bosnian"},
		{Code: "bg-BG", Title: "Bulgarian"},
		{Code: "ca-ES", Title: "Catalan"},
		{Code: "zh-CN", Title: "Chinese"},
		{Code: "hr-HR", Title: "Croatian"},
		{Code: "cs-CZ", Title: "Czech"},
		{Code: "da-DK", Title: "Danish"},
		{Code: "nl-NL", Title: "Dutch"},
		{Code: "en-US", Title: "English"},
		{Code: "et-EE", Title: "Estonian"},
		{Code: "fi-FI", Title: "Finnish"},
		{Code: "fr-FR", Title: "French"},
		{Code: "gl-ES", Title: "Galician"},
		{Code: "de-DE", Title: "German"},
		{Code: "el-GR", Title: "Greek"},
		{Code: "he-IL", Title: "Hebrew"},
		{Code: "hi-IN", Title: "Hindi"},
		{Code: "hu-HU", Title: "Hungarian"},
		{Code: "is-IS", Title: "Icelandic"},
		{Code: "id-ID", Title: "Indonesian"},
		{Code: "it-IT", Title: "Italian"},
		{Code: "ja-JP", Title: "Japanese"},
		{Code: "kn-IN", Title: "Kannada"},
		{Code: "kk-KZ", Title: "Kazakh"},
		{Code: "ko-KR", Title: "Korean"},
		{Code: "lv-LV", Title: "Latvian"},
		{Code: "lt-LT", Title: "Lithuanian"},
		{Code: "mk-MK", Title: "Macedonian"},
		{Code: "ms-MY", Title: "Malay"},
		{Code: "mr-IN", Title: "Marathi"},
		{Code: "mi-NZ", Title: "Maori"},
		{Code: "ne-NP", Title: "Nepali"},
		{Code: "no-NO", Title: "Norwegian"},
		{Code: "fa-IR", Title: "Persian"},
		{Code: "pl-PL", Title: "Polish"},
		{Code: "pt-PT", Title: "Portuguese"},
		{Code: "ro-RO", Title: "Romanian"},
		{Code: "ru-RU", Title: "Russian"},
		{Code: "sr-RS", Title: "Serbian"},
		{Code: "sk-SK", Title: "Slovak"},
		{Code: "sl-SI", Title: "Slovenian"},
		{Code: "es-ES", Title: "Spanish"},
		{Code: "sw-KE", Title: "Swahili"},
		{Code: "sv-SE", Title: "Swedish"},
		{Code: "tl-PH", Title: "Tagalog"},
		{Code: "ta-IN", Title: "Tamil"},
		{Code: "th-TH", Title: "Thai"},
		{Code: "tr-TR", Title: "Turkish"},
		{Code: "uk-UA", Title: "Ukrainian"},
		{Code: "ur-PK", Title: "Urdu"},
		{Code: "vi-VN", Title: "Vietnamese"},
		{Code: "cy-GB", Title: "Welsh"},
	}

	// Text to speech - OpenAI API
	// https://platform.openai.com/docs/guides/text-to-speech#supported-output-formats
	formats = []types.VoiceFormat{
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg"},
		{Name: "Opus", Extension: ".opus", MimeType: "audio/opus"},
		{Name: "AAC", Extension: ".aac", MimeType: "audio/aac"},
		{Name: "FLAC", Extension: ".flac", MimeType: "audio/flac"},
		{Name: "WAV", Extension: ".wav", MimeType: "audio/wav"},
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm"},
	}
)

func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	return mo.Ok[any](types.ListVoicesResponse{
		Voices: []types.Voice{
			{
				ID:                "alloy",
				Name:              "Alloy",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/alloy.wav",
			},
			{
				ID:                "ash",
				Name:              "Ash",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/ash.wav",
			},
			{
				ID:                "coral",
				Name:              "Coral",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/coral.wav",
			},
			{
				ID:                "echo",
				Name:              "Echo",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/echo.wav",
			},
			{
				ID:                "fable",
				Name:              "Fable",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/fable.wav",
			},
			{
				ID:                "onyx",
				Name:              "Onyx",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/onyx.wav",
			},
			{
				ID:                "nova",
				Name:              "Nova",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/nova.wav",
			},
			{
				ID:                "sage",
				Name:              "Sage",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/sage.wav",
			},
			{
				ID:                "shimmer",
				Name:              "Shimmer",
				Description:       "",
				Labels:            map[string]any{},
				Tags:              make([]string, 0),
				Languages:         languages,
				Formats:           formats,
				CompatibleModels:  []string{"tts-1", "tts-1-hd"},
				PredefinedOptions: nil,
				PreviewAudioURL:   "https://cdn.openai.com/API/docs/audio/shimmer.wav",
			},
		},
	})
}
