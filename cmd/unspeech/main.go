package main

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"

	"github.com/moeru-ai/unspeech/internal/middlewares"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "unspeech",
		RunE: func(cmd *cobra.Command, args []string) error {
			e := echo.New()

			e.Use(middlewares.CORS())
			e.Use(middlewares.HandleErrors())

			e.POST("/api/v1/elevenlabs/tts", func(c echo.Context) error {
				return apierrors.NewErrInternal().WithCaller()
			})

			e.RouteNotFound("/*", middlewares.NotFound)

			return e.Start(":8080")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
