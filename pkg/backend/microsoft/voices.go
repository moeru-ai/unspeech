package microsoft

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/nekomeowww/xo"
	"github.com/samber/mo"
)

type VoiceTagKey string

const (
	VoiceTagKeyTailoredScenarios  VoiceTagKey = "TailoredScenarios"
	VoiceTagKeyVoicePersonalities VoiceTagKey = "VoicePersonalities"
)

var (
	// https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech
	formats = []types.VoiceFormat{
		// AMR-WB
		{Name: "AMR-WB", Extension: ".amr", MimeType: "audio/amr-wb", SampleRate: 16000, FormatCode: "amr-wb-16000hz"}, //nolint:mnd

		// Opus formats
		{Name: "Opus", Extension: ".opus", MimeType: "audio/opus", SampleRate: 16000, Bitrate: 32, FormatCode: "audio-16khz-16bit-32kbps-mono-opus"}, //nolint:mnd
		{Name: "Opus", Extension: ".opus", MimeType: "audio/opus", SampleRate: 24000, Bitrate: 24, FormatCode: "audio-24khz-16bit-24kbps-mono-opus"}, //nolint:mnd
		{Name: "Opus", Extension: ".opus", MimeType: "audio/opus", SampleRate: 24000, Bitrate: 48, FormatCode: "audio-24khz-16bit-48kbps-mono-opus"}, //nolint:mnd

		// MP3 formats
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 16000, Bitrate: 32, FormatCode: "audio-16khz-32kbitrate-mono-mp3"},   //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 16000, Bitrate: 64, FormatCode: "audio-16khz-64kbitrate-mono-mp3"},   //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 16000, Bitrate: 128, FormatCode: "audio-16khz-128kbitrate-mono-mp3"}, //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 24000, Bitrate: 48, FormatCode: "audio-24khz-48kbitrate-mono-mp3"},   //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 24000, Bitrate: 96, FormatCode: "audio-24khz-96kbitrate-mono-mp3"},   //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 24000, Bitrate: 160, FormatCode: "audio-24khz-160kbitrate-mono-mp3"}, //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 48000, Bitrate: 96, FormatCode: "audio-48khz-96kbitrate-mono-mp3"},   //nolint:mnd
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 48000, Bitrate: 192, FormatCode: "audio-48khz-192kbitrate-mono-mp3"}, //nolint:mnd

		// G722
		{Name: "G722", Extension: ".g722", MimeType: "audio/g722", SampleRate: 16000, Bitrate: 64, FormatCode: "g722-16khz-64kbps"}, //nolint:mnd

		// Ogg Opus formats
		{Name: "Ogg Opus", Extension: ".ogg", MimeType: "audio/ogg", SampleRate: 16000, FormatCode: "ogg-16khz-16bit-mono-opus"}, //nolint:mnd
		{Name: "Ogg Opus", Extension: ".ogg", MimeType: "audio/ogg", SampleRate: 24000, FormatCode: "ogg-24khz-16bit-mono-opus"}, //nolint:mnd
		{Name: "Ogg Opus", Extension: ".ogg", MimeType: "audio/ogg", SampleRate: 48000, FormatCode: "ogg-48khz-16bit-mono-opus"}, //nolint:mnd

		// Raw PCM formats
		{Name: "PCM A-law", Extension: ".alaw", MimeType: "audio/x-alaw-basic", SampleRate: 8000, FormatCode: "raw-8khz-8bit-mono-alaw"},   //nolint:mnd
		{Name: "PCM μ-law", Extension: ".ulaw", MimeType: "audio/x-mulaw-basic", SampleRate: 8000, FormatCode: "raw-8khz-8bit-mono-mulaw"}, //nolint:mnd
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 8000, FormatCode: "raw-8khz-16bit-mono-pcm"},                   //nolint:mnd
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 16000, FormatCode: "raw-16khz-16bit-mono-pcm"},                 //nolint:mnd
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 22050, FormatCode: "raw-22050hz-16bit-mono-pcm"},               //nolint:mnd
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 24000, FormatCode: "raw-24khz-16bit-mono-pcm"},                 //nolint:mnd
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 44100, FormatCode: "raw-44100hz-16bit-mono-pcm"},               //nolint:mnd
		{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 48000, FormatCode: "raw-48khz-16bit-mono-pcm"},                 //nolint:mnd

		// TrueSilk formats
		{Name: "TrueSilk", Extension: ".silk", MimeType: "audio/silk", SampleRate: 16000, FormatCode: "raw-16khz-16bit-mono-truesilk"}, //nolint:mnd
		{Name: "TrueSilk", Extension: ".silk", MimeType: "audio/silk", SampleRate: 24000, FormatCode: "raw-24khz-16bit-mono-truesilk"}, //nolint:mnd

		// WebM formats
		{Name: "WebM Opus", Extension: ".webm", MimeType: "audio/webm", SampleRate: 16000, FormatCode: "webm-16khz-16bit-mono-opus"},                     //nolint:mnd
		{Name: "WebM Opus", Extension: ".webm", MimeType: "audio/webm", SampleRate: 24000, Bitrate: 24, FormatCode: "webm-24khz-16bit-24kbps-mono-opus"}, //nolint:mnd
		{Name: "WebM Opus", Extension: ".webm", MimeType: "audio/webm", SampleRate: 24000, FormatCode: "webm-24khz-16bit-mono-opus"},                     //nolint:mnd
	}
)

type Voice struct {
	Name            string                `json:"Name"`
	DisplayName     string                `json:"DisplayName"`
	LocalName       string                `json:"LocalName"`
	ShortName       string                `json:"ShortName"`
	Gender          string                `json:"Gender"`
	Locale          string                `json:"Locale"`
	LocaleName      string                `json:"LocaleName"`
	SampleRateHertz string                `json:"SampleRateHertz"`
	VoiceType       string                `json:"VoiceType"`
	Status          string                `json:"Status"`
	VoiceTag        map[VoiceTagKey][]any `json:"VoiceTag"`
	WordsPerMinute  string                `json:"WordsPerMinute"`
}

func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	region := options.MustGet().ExtraQuery.Get("region")
	if region == "" {
		region = "eastasia"
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, "https://"+region+".tts.speech.microsoft.com/cognitiveservices/voices/list", nil)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithError(err).WithCaller())
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithError(err).WithCaller())
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 400 && res.StatusCode < 600 {
		switch {
		case strings.HasPrefix(res.Header.Get("Content-Type"), "application/json"):
			return mo.Err[any](apierrors.
				NewUpstreamError(res.StatusCode).
				WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()))
		case strings.HasPrefix(res.Header.Get("Content-Type"), "text/"):
			return mo.Err[any](apierrors.
				NewUpstreamError(res.StatusCode).
				WithDetail(utils.NewTextResponseError(res.StatusCode, res.Body).OrEmpty().Error()))
		default:
			slog.Warn("unknown upstream error with unknown Content-Type",
				slog.Int("status", res.StatusCode),
				slog.String("content_type", res.Header.Get("Content-Type")),
				slog.String("content_length", res.Header.Get("Content-Length")),
			)
		}
	}

	var response []Voice
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithError(err).WithCaller())
	}

	voices := make([]types.Voice, 0, len(response))

	for _, voice := range response {
		tags := make([]string, 0, len(voice.VoiceTag[VoiceTagKeyTailoredScenarios])+len(voice.VoiceTag[VoiceTagKeyVoicePersonalities]))

		for _, tag := range voice.VoiceTag[VoiceTagKeyTailoredScenarios] {
			tags = append(tags, xo.Stringify(tag))
		}

		for _, tag := range voice.VoiceTag[VoiceTagKeyVoicePersonalities] {
			tags = append(tags, xo.Stringify(tag))
		}

		voices = append(voices, types.Voice{
			ID:          voice.ShortName,
			Name:        voice.DisplayName,
			Description: voice.Name,
			Labels: map[string]any{
				types.VoiceLabelKeyType:   voice.VoiceType,
				types.VoiceLabelKeyAccent: voice.LocalName,
				types.VoiceLabelKeyGender: voice.Gender,
				"tailoredScenarios":       voice.VoiceTag[VoiceTagKeyTailoredScenarios],
				"voicePersonalities":      voice.VoiceTag[VoiceTagKeyVoicePersonalities],
			},
			Tags:             tags,
			Formats:          formats,
			CompatibleModels: []string{"v1"},
			PreviewAudioURL:  "https://staticassets.api.speech.microsoft.com/blobs/ttsvoice/Masterpieces/" + voice.Locale + "-" + voice.DisplayName + "-General-Audio.wav",
			Languages: []types.VoiceLanguage{
				{
					Title: voice.LocalName,
					Code:  voice.Locale,
				},
			},
		})
	}

	return mo.Ok[any](types.ListVoicesResponse{
		Voices: voices,
	})
}
