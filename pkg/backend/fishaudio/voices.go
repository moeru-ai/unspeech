package fishaudio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const defaultVoicesURL = "https://api.fish.audio/model"

// Model generations selectable via the `model` header.
// https://docs.fish.audio/api-reference/endpoint/openapi-v1/text-to-speech
const (
	modelS1        = "s1"
	modelS2Pro     = "s2-pro"
	modelS21Pro    = "s2.1-pro"
	modelS21ProFre = "s2.1-pro-free"
)

var compatibleModels = []string{modelS1, modelS2Pro, modelS21Pro, modelS21ProFre}

// languageTitles maps Fish Audio's bare language codes to display titles.
// Codes missing here fall back to the raw code.
var languageTitles = map[string]string{
	"en": "English",
	"zh": "Chinese",
	"ja": "Japanese",
	"ko": "Korean",
	"de": "German",
	"fr": "French",
	"es": "Spanish",
	"ar": "Arabic",
	"ru": "Russian",
	"nl": "Dutch",
	"it": "Italian",
	"pl": "Polish",
	"pt": "Portuguese",
}

var formats = []types.VoiceFormat{
	{Name: "MP3", Extension: ".mp3", MimeType: "audio/mpeg", SampleRate: 44100, FormatCode: "mp3"},    //nolint:mnd
	{Name: "WAV", Extension: ".wav", MimeType: "audio/wav", SampleRate: 44100, FormatCode: "wav"},     //nolint:mnd
	{Name: "Opus", Extension: ".opus", MimeType: "audio/opus", SampleRate: 48000, FormatCode: "opus"}, //nolint:mnd
	{Name: "PCM", Extension: ".pcm", MimeType: "audio/L16", SampleRate: 44100, FormatCode: "pcm"},     //nolint:mnd
}

type voiceModelSample struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

type voiceModelAuthor struct {
	ID       string `json:"_id"`
	Nickname string `json:"nickname"`
}

// voiceModel is a Fish Audio community/user voice model ("reference"),
// whose _id is used as reference_id in TTS requests.
type voiceModel struct {
	ID          string             `json:"_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Languages   []string           `json:"languages"`
	Tags        []string           `json:"tags"`
	Samples     []voiceModelSample `json:"samples"`
	Author      voiceModelAuthor   `json:"author"`
	LikeCount   int                `json:"like_count"`
	TaskCount   int                `json:"task_count"`
}

type listVoiceModelsResponse struct {
	Items []voiceModel `json:"items"`
	Total int          `json:"total"`
}

// VoicesCredentials holds the Fish Audio API key for voices listing.
type VoicesCredentials struct {
	// APIKey is sent upstream as a Bearer token.
	APIKey string
}

// UpstreamError carries a non-2xx response body from the Fish Audio API.
type UpstreamError struct {
	StatusCode  int
	ContentType string
	Body        string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("fish audio returned %d: %s", e.StatusCode, e.Body)
}

// ListVoices fetches voice models from Fish Audio and maps them to types.Voice.
//
// Callers can narrow results with `query` (already validated url.Values):
// title, language, self, sort_by, page_size, and page_number are forwarded.
func ListVoices(ctx context.Context, creds VoicesCredentials, query url.Values) ([]types.Voice, error) {
	reqURL := lo.Must(url.Parse(defaultVoicesURL))

	upstreamQuery := url.Values{}
	// Popular voices first gives a useful default picker ordering; the
	// upstream default (recently created) surfaces mostly untested models.
	upstreamQuery.Set("sort_by", "task_count")
	upstreamQuery.Set("page_size", "50")

	for _, key := range []string{"title", "language", "self", "sort_by", "page_size", "page_number"} {
		if value := query.Get(key); value != "" {
			upstreamQuery.Set(key, value)
		}
	}

	reqURL.RawQuery = upstreamQuery.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("fishaudio: build request: %w", err)
	}

	if creds.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fishaudio: call voices: %w", err)
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 400 && res.StatusCode < 600 {
		body, _ := io.ReadAll(res.Body)

		return nil, &UpstreamError{
			StatusCode:  res.StatusCode,
			ContentType: res.Header.Get("Content-Type"),
			Body:        string(body),
		}
	}

	var response listVoiceModelsResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("fishaudio: decode voices: %w", err)
	}

	return lo.Map(response.Items, func(item voiceModel, _ int) types.Voice {
		return mapVoice(item)
	}), nil
}

func mapVoice(item voiceModel) types.Voice {
	voice := types.Voice{
		ID:          item.ID,
		Name:        item.Title,
		Description: item.Description,
		Labels: map[string]any{
			"author":     item.Author.Nickname,
			"like_count": item.LikeCount,
			"task_count": item.TaskCount,
		},
		Tags:             item.Tags,
		Languages:        make([]types.VoiceLanguage, 0, len(item.Languages)),
		Formats:          formats,
		CompatibleModels: compatibleModels,
	}

	for _, code := range item.Languages {
		title, ok := languageTitles[code]
		if !ok {
			title = code
		}

		voice.Languages = append(voice.Languages, types.VoiceLanguage{Code: code, Title: title})
	}

	if len(item.Samples) > 0 {
		voice.PreviewAudioURL = item.Samples[0].Audio
	}

	return voice
}

func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	creds := VoicesCredentials{
		APIKey: strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "),
	}

	query := url.Values{}
	if opts, ok := options.Get(); ok && opts.ExtraQuery != nil {
		query = opts.ExtraQuery
	}

	voices, err := ListVoices(c.Request().Context(), creds, query)
	if err != nil {
		var upErr *UpstreamError
		if errors.As(err, &upErr) {
			switch {
			case strings.HasPrefix(upErr.ContentType, "application/json"):
				return mo.Err[any](apierrors.
					NewUpstreamError(upErr.StatusCode).
					WithDetail(utils.NewJSONResponseError(upErr.StatusCode, strings.NewReader(upErr.Body)).OrEmpty().Error()))
			case strings.HasPrefix(upErr.ContentType, "text/"):
				return mo.Err[any](apierrors.
					NewUpstreamError(upErr.StatusCode).
					WithDetail(utils.NewTextResponseError(upErr.StatusCode, strings.NewReader(upErr.Body)).OrEmpty().Error()))
			default:
				slog.Warn("unknown upstream error with unknown Content-Type",
					slog.Int("status", upErr.StatusCode),
					slog.String("content_type", upErr.ContentType),
				)

				return mo.Err[any](apierrors.NewUpstreamError(upErr.StatusCode).WithDetail(upErr.Body))
			}
		}

		return mo.Err[any](apierrors.NewErrBadGateway().WithError(err).WithCaller())
	}

	return mo.Ok[any](types.ListVoicesResponse{Voices: voices})
}
