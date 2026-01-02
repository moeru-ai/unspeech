package deepgram

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/mo"
)

var (
	formats = []types.VoiceFormat{
		{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg"},
		{Name: "WAV", Extension: ".wav", MimeType: "audio/wav"},
		{Name: "FLAC", Extension: ".flac", MimeType: "audio/flac"},
		{Name: "AAC", Extension: ".aac", MimeType: "audio/aac"},
		{Name: "OPUS", Extension: ".opus", MimeType: "audio/opus"},
	}
)

type DeepgramModel struct {
	Name          string   `json:"name"`
	CanonicalName string   `json:"canonical_name"`
	Architecture  string   `json:"architecture"`
	Languages     []string `json:"languages"`
	Version       string   `json:"version"`
	UUID          string   `json:"uuid"`
}

type DeepgramModelsResponse struct {
	TTS []DeepgramModel `json:"tts"`
}

func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	// Deepgram requires authentication to list models
	auth := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		auth = "Token " + strings.TrimPrefix(auth, "Bearer ")
	}

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodGet,
		"https://api.deepgram.com/v1/models",
		nil,
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](
			apierrors.NewErrBadGateway().
				WithDetail(err.Error()).
				WithError(err).
				WithCaller(),
		)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		return mo.Err[any](
			apierrors.NewUpstreamError(res.StatusCode).
				WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
		)
	}

	var response DeepgramModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("failed to decode upstream response"))
	}

	voices := make([]types.Voice, 0, len(response.TTS))
	for _, model := range response.TTS {
		// Map languages
		langs := make([]types.VoiceLanguage, len(model.Languages))
		for i, code := range model.Languages {
			langs[i] = types.VoiceLanguage{
				Code:  code,
				Title: code, // Deepgram doesn't provide human-readable titles in the API response
			}
		}

		voices = append(voices, types.Voice{
			ID:               model.CanonicalName,
			Name:             model.Name,
			Languages:        langs,
			Formats:          formats,
			CompatibleModels: []string{model.Architecture}, // e.g., "aura"
			Tags:             make([]string, 0),
			Description:      "Deepgram " + model.Architecture + " voice",
		})
	}

	return mo.Ok[any](types.ListVoicesResponse{
		Voices: voices,
	})
}
