package elevenlabs

import (
	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/samber/mo"
)

func HandleVoices(c echo.Context, options mo.Option[types.VoicesRequestOptions]) mo.Result[any] {
	return mo.Err[any](apierrors.NewErrNotFound().WithDetail("not implemented"))
}
