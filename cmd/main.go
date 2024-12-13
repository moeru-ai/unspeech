package main

import (
	"github.com/labstack/echo/v4"
	"github.com/samber/mo"
	"github.com/spf13/cobra"

	"github.com/moeru-ai/unspeech/internal/middlewares"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/ho"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "unspeech",
		RunE: func(cmd *cobra.Command, args []string) error {
			e := echo.New()

			e.Use(middlewares.CORS())
			e.Use(middlewares.HandleErrors())

			e.POST("/api/v1/elevenlabs/tts", ho.MonadEcho1(func(c echo.Context) mo.Result[any] {
				return mo.Err[any](apierrors.NewErrInternal().WithCaller())
			}))
			e.RouteNotFound("/*", ho.MonadEcho1(middlewares.NotFound))

			return e.Start(":5933")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
