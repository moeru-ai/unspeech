package ho

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func MonadEcho1[T any](cb func(c echo.Context) mo.Result[T]) echo.HandlerFunc {
	return func(c echo.Context) error {
		res := cb(c)
		if res.IsError() {
			return res.Error()
		}

		if lo.IsNil(res.MustGet()) {
			return nil
		}

		return c.JSON(http.StatusOK, res.MustGet())
	}
}
