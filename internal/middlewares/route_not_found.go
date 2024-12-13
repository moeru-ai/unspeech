package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/mo"
)

func NotFound(c echo.Context) mo.Result[any] {
	return mo.Err[any](apierrors.NewErrNotFound())
}
