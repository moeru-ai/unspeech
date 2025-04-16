package volcengine

import (
	_ "embed"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/samber/mo"
)

var (
	//go:embed voices.json
	voicesJSON string
)

type VoicesResponseDataResourcePackDetails struct {
	DemoLink            string `json:"demo_link"`
	Language            string `json:"language"`
	RecommendedScenario string `json:"recommended_scenario"`
	ToneNumber          string `json:"tone_number"`
	VoiceType           string `json:"voice_type"`
}

type VoicesResponseDataResourcePack struct {
	InstanceNumber    string                                `json:"instance_number"`
	IsShareable       bool                                  `json:"is_shareable"`
	ResourceID        string                                `json:"resource_id"`
	Code              string                                `json:"code"`
	ConfigurationCode string                                `json:"configuration_code"`
	ResourceDisplay   string                                `json:"resource_display"`
	RawType           string                                `json:"raw_type"`
	Type              string                                `json:"type"`
	PurchasedAmount   string                                `json:"purchased_amount"`
	CurrentUsage      string                                `json:"current_usage"`
	Expires           string                                `json:"expires"`
	Details           VoicesResponseDataResourcePackDetails `json:"details"`
	GroupName         string                                `json:"group_name"`
	Alias             string                                `json:"alias"`
	TrainID           string                                `json:"train_id"`
	State             string                                `json:"state"`
}

type VoicesResponseData struct {
	ResourcePacks []VoicesResponseDataResourcePack `json:"resource_packs"`
}

type VoicesResponse struct {
	Status string             `json:"status"`
	Error  any                `json:"error,omitempty"`
	Data   VoicesResponseData `json:"data"`
}

func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	var voicesData VoicesResponse

	err := json.Unmarshal([]byte(voicesJSON), &voicesData)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithCaller())
	}

	voices := make([]types.Voice, 0, len(voicesData.Data.ResourcePacks))

	for _, voice := range voicesData.Data.ResourcePacks {
		voices = append(voices, types.Voice{
			ID:          voice.Code,
			Name:        voice.ResourceDisplay,
			Description: voice.Details.Language + " " + voice.Details.RecommendedScenario,
			Labels: map[string]any{
				types.VoiceLabelKeyType:   voice.Details.VoiceType,
				types.VoiceLabelKeyAccent: voice.Details.Language,
				"configuration_code":      voice.ConfigurationCode,
				"tone_number":             voice.Details.ToneNumber,
				"tailoredScenarios":       voice.Details.RecommendedScenario,
			},
			Tags: make([]string, 0),
			Formats: []types.VoiceFormat{
				// https://www.volcengine.com/docs/6561/1257584
				{Name: "WAV", Extension: ".wav", MimeType: "audio/wav", SampleRate: 24000, Bitrate: 16, FormatCode: "wav"},         //nolint:mnd
				{Name: "PCM", Extension: ".pcm", MimeType: "audio/pcm", SampleRate: 24000, Bitrate: 16, FormatCode: "pcm"},         //nolint:mnd
				{Name: "Opus", Extension: ".opus", MimeType: "audio/opus", SampleRate: 24000, Bitrate: 16, FormatCode: "ogg_opus"}, //nolint:mnd
				{Name: "MP3", Extension: ".mp3", MimeType: "audio/mp3", SampleRate: 24000, Bitrate: 16, FormatCode: "mp3"},         //nolint:mnd
			},
			CompatibleModels: []string{"v1"},
			PreviewAudioURL:  voice.Details.DemoLink,
			Languages: []types.VoiceLanguage{
				{
					Title: voice.Details.Language,
					Code:  voice.Details.Language,
				},
			},
		})
	}

	return mo.Ok[any](types.ListVoicesResponse{
		Voices: voices,
	})
}
