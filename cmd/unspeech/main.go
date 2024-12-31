package main

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/lmittmann/tint"
	slogecho "github.com/samber/slog-echo"
	"github.com/spf13/cobra"

	"github.com/moeru-ai/unspeech/internal/middlewares"
	"github.com/moeru-ai/unspeech/pkg/backend"
	"github.com/moeru-ai/unspeech/pkg/ho"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "unspeech",
		// TODO: set version
		Version: "0.0.0",
		RunE: func(cmd *cobra.Command, args []string) error {
			e := echo.New()

			e.HideBanner = true

			e.Use(slogecho.New(slog.New(tint.NewHandler(os.Stdout, nil))))
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
