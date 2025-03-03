package backend

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/mo"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/microsoft"
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
		return openai(c, utils.ResultToOption(options))
	case "elevenlabs":
		return elevenlabs(c, utils.ResultToOption(options))
	case "koemotion":
		return koemotion(c, utils.ResultToOption(options))
	case "microsoft", "azure":
		return microsoft.Handle(c, utils.ResultToOption(options))
	default:
		return mo.Err[any](apierrors.NewErrBadRequest().WithDetail("unsupported backend"))
	}
}
