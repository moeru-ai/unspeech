package backend

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
)

// https://platform.openai.com/docs/api-reference/audio/createSpeech
type Options struct {
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
}

type FullOptions struct {
	Options
	Backend string `json:"backend"`
	Model   string `json:"model"`
}

func Speech(c echo.Context) mo.Result[any] {
	options := new(Options)

	if err := c.Bind(options); err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	if options.Model == "" || options.Input == "" || options.Voice == "" {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	backendAndModel := lo.Ternary(
		strings.Contains(options.Model, ":"),
		//nolint:mnd
		strings.SplitN(options.Model, ":", 2),
		[]string{options.Model, ""},
	)

	fullOptions := FullOptions{
		Options: *options,
		Backend: backendAndModel[0],
		Model:   backendAndModel[1],
	}

	return openai(c, fullOptions)
	// return mo.Ok[any](c.JSON(http.StatusOK, fullOptions))
}
