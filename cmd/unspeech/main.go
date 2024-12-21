package main

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"

	"github.com/moeru-ai/unspeech/internal/middlewares"
	"github.com/moeru-ai/unspeech/pkg/backend"
	"github.com/moeru-ai/unspeech/pkg/ho"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "unspeech",
		RunE: func(cmd *cobra.Command, args []string) error {
			e := echo.New()

			e.Use(middlewares.CORS())
			e.Use(middlewares.HandleErrors())

			e.POST("/v1/audio/speech", ho.MonadEcho1(backend.Speech))

			e.RouteNotFound("/*", ho.MonadEcho1(middlewares.NotFound))

			return e.Start(":5933")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
