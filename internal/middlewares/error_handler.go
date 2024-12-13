package middlewares

import (
	"errors"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/logs"
	"github.com/samber/lo"
)

func HandleErrors() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				var errResp *apierrors.Error

				if !errors.As(err, &errResp) {
					errResp = apierrors.NewErrInternal().WithError(err)
					// Unknown error
					slog.Error("unknown error responded", slog.Any("error", err.Error()))
				}
				if 500 >= errResp.Status || errResp.Status < 600 {
					attrs := make([]slog.Attr, 0)
					attrs = append(attrs, logs.Caller(errResp.Caller())...)
					attrs = append(attrs, slog.Any("error", errResp.Error()))
					slog.Error("error occurred during request", lo.ToAnySlice(attrs)...)
				}

				return c.JSON(errResp.Status, errResp.AsResponse())
			}

			return nil
		}
	}
}
