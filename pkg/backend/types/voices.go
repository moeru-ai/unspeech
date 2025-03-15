package types

import (
	"net/http"
	"net/url"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/mo"
)

type VoicesRequestOptions struct {
	Backend string `json:"provider"`

	ExtraQuery url.Values `json:"extra_query"`
}

func NewVoicesRequestOptions(request *http.Request) mo.Result[VoicesRequestOptions] {
	provider := request.URL.Query().Get("provider")
	if provider == "" {
		return mo.Err[VoicesRequestOptions](
			apierrors.
				NewErrInvalidArgument().
				WithDetail("provider is required").
				WithSourceParameter("provider"),
		)
	}

	return mo.Ok(VoicesRequestOptions{
		Backend:    provider,
		ExtraQuery: request.URL.Query(),
	})
}

type VoiceCommonLabels struct {
	Gender       string `json:"gender"`
	Language     string `json:"language"`
	LanguageCode string `json:"language_code"`
	SampleRate   int    `json:"sample_rate"`
}

type VoiceLabelKey = string

const (
	VoiceLabelKeyAge    VoiceLabelKey = "age"
	VoiceLabelKeyGender VoiceLabelKey = "gender"
	VoiceLabelKeyAccent VoiceLabelKey = "accent"
	VoiceLabelKeyType   VoiceLabelKey = "type"
)

type VoiceLanguage struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

type VoiceFormat struct {
	Name       string `json:"name"`
	Extension  string `json:"extension"`
	MimeType   string `json:"mime_type"`
	SampleRate int    `json:"sample_rate,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"`
	FormatCode string `json:"format_code,omitempty"` // For provider-specific format codes
}

type Voice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Labels map[string]any `json:"labels"`
	Tags   []string       `json:"tags"`

	Languages         []VoiceLanguage `json:"languages"`
	Formats           []VoiceFormat   `json:"formats"`
	CompatibleModels  []string        `json:"compatible_models"`
	PredefinedOptions map[string]any  `json:"predefined_options"`
	PreviewAudioURL   string          `json:"preview_audio_url"`
}

type ListVoicesResponse struct {
	Voices []Voice `json:"voices"`
}
