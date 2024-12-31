package backend

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/utils"
)

// OpenAISpeechRequestOptions represent API parameters refer to https://platform.openai.com/docs/api-reference/audio/createSpeech
type OpenAISpeechRequestOptions struct {
	// (required) One of the available TTS models.
	Model string `json:"model"`
	// (required) The text to generate audio for.
	Input string `json:"input"`
	// (required) The voice to use when generating the audio.
	Voice string `json:"voice"`

	// The format to audio in.
	// Supported formats are mp3, opus, aac, flac, wav, and pcm.
	// mp3 is the default.
	ResponseFormat string `json:"response_format,omitempty"`
	// The speed of the generated audio.
	// Select a value from 0.25 to 4.0.
	// 1.0 is the default.
	Speed int `json:"speed,omitempty"`

	// Extension: allows you to add custom content to body.
	ExtraBody map[string]any `json:"extra_body,omitempty"`
}

type SpeechRequestOptions struct {
	OpenAISpeechRequestOptions

	Backend string `json:"backend"`
	Model   string `json:"model"`

	body          mo.Option[*bytes.Buffer]
	bodyParsedMap map[string]any
}

func (o SpeechRequestOptions) AsBuffer() mo.Option[*bytes.Buffer] {
	return o.body
}

func (o SpeechRequestOptions) AsMap() map[string]any {
	return o.bodyParsedMap
}

func NewSpeechRequestOptions(body io.ReadCloser) mo.Result[SpeechRequestOptions] {
	buffer := new(bytes.Buffer)

	_, err := buffer.ReadFrom(body)
	if err != nil {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrBadRequest().WithDetail(err.Error()))
	}

	var optionsMap map[string]any

	err = json.Unmarshal(buffer.Bytes(), &optionsMap)
	if err != nil {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrBadRequest().WithDetail(err.Error()))
	}

	var options OpenAISpeechRequestOptions

	err = json.Unmarshal(buffer.Bytes(), &options)
	if err != nil {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrBadRequest().WithDetail(err.Error()))
	}

	if options.Model == "" || options.Input == "" || options.Voice == "" {
		return mo.Err[SpeechRequestOptions](apierrors.NewErrInvalidArgument().WithDetail("either one of model, input, and voice parameter is required"))
	}

	backendAndModel := lo.Ternary(
		strings.Contains(options.Model, "/"),
		strings.SplitN(options.Model, "/", 2), //nolint:mnd
		[]string{options.Model, ""},
	)

	return mo.Ok(SpeechRequestOptions{
		OpenAISpeechRequestOptions: options,
		Backend:                    backendAndModel[0],
		Model:                      backendAndModel[1],
		body:                       mo.Some(buffer),
		bodyParsedMap:              optionsMap,
	})
}

func Speech(c echo.Context) mo.Result[any] {
	options := NewSpeechRequestOptions(c.Request().Body)
	if options.IsError() {
		return mo.Err[any](options.Error())
	}

	switch options.MustGet().Backend {
	case "openai":
		return openai(c, utils.ResultToOption(options))
	case "elevenlabs":
		return elevenlabs(c, utils.ResultToOption(options))
	case "koemotion":
		return koemotion(c, utils.ResultToOption(options))
	default:
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported backend"))
	}
}
