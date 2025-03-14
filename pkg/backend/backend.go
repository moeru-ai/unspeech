package backend

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/elevenlabs"
	"github.com/moeru-ai/unspeech/pkg/backend/koemotion"
	"github.com/moeru-ai/unspeech/pkg/backend/microsoft"
	"github.com/moeru-ai/unspeech/pkg/backend/openai"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
)

func Speech(c echo.Context) mo.Result[any] {
	options := types.NewSpeechRequestOptions(c.Request().Body)
	if options.IsError() {
		return mo.Err[any](options.Error())
	}

	switch options.MustGet().Backend {
	case "openai":
		return openai.HandleSpeech(c, utils.ResultToOption(options))
	case "elevenlabs":
		return elevenlabs.HandleSpeech(c, utils.ResultToOption(options))
	case "koemotion":
		return koemotion.HandleSpeech(c, utils.ResultToOption(options))
	case "microsoft", "azure":
		return microsoft.HandleSpeech(c, utils.ResultToOption(options))
	default:
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported backend"))
	}
}

func Voices(c echo.Context) mo.Result[any] {
	options := types.NewVoicesRequestOptions(c.Request())
	if options.IsError() {
		return mo.Err[any](options.Error())
	}

	switch options.MustGet().Backend {
	case "openai":
		return openai.HandleVoices(c, utils.ResultToOption(options))
	case "elevenlabs":
		return elevenlabs.HandleVoices(c, utils.ResultToOption(options))
	case "koemotion":
		return koemotion.HandleVoices(c, utils.ResultToOption(options))
	case "microsoft", "azure":
		return microsoft.HandleVoices(c, utils.ResultToOption(options))
	default:
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported backend"))
	}
}
