package backend

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
)

// Options represent API parameters refer to https://platform.openai.com/docs/api-reference/audio/createSpeech
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
	var options Options

	if err := c.Bind(&options); err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest())
	}
	if options.Model == "" || options.Input == "" || options.Voice == "" {
		return mo.Err[any](apierrors.NewErrInvalidArgument().WithDetail("either one of model, input, and voice parameter is required"))
	}

	backendAndModel := lo.Ternary(
		strings.Contains(options.Model, "/"),
		strings.SplitN(options.Model, "/", 2), //nolint:mnd
		[]string{options.Model, ""},
	)

	fullOptions := FullOptions{
		Options: options,
		Backend: backendAndModel[0],
		Model:   backendAndModel[1],
	}

	switch backendAndModel[0] {
	case "openai":
		return openai(c, fullOptions)
	case "elevenlabs":
		return elevenlabs(c, fullOptions)
	default:
		return mo.Err[any](apierrors.NewErrBadRequest())
	}
}
