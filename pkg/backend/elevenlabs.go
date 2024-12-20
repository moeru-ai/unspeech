package backend

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request
type ElevenLabsOptions struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id,omitempty"`
}

func elevenlabs(c echo.Context, options FullOptions) mo.Result[any] {
	reqURL := lo.Must(url.Parse("https://api.elevenlabs.io/v1/text-to-speech")).
		JoinPath(options.Voice).
		String()

	values := ElevenLabsOptions{
		Text:    options.Input,
		ModelID: options.Model,
	}

	payload := lo.Must(json.Marshal(values))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		reqURL,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	// Rewrite the Authorization header
	//nolint:canonicalheader
	req.Header.Set("xi-api-key", strings.TrimPrefix(
		c.Request().Header.Get("Authorization"),
		"Bearer ",
	))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadGateway().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 400 && res.StatusCode < 600 {
		switch {
		case strings.HasPrefix(res.Header.Get("Content-Type"), "application/json"):
			return mo.Err[any](apierrors.
				NewUpstreamError(res.StatusCode).
				WithDetail(NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()))
		case strings.HasPrefix(res.Header.Get("Content-Type"), "text/"):
			return mo.Err[any](apierrors.
				NewUpstreamError(res.StatusCode).
				WithDetail(NewTextResponseError(res.StatusCode, res.Body).OrEmpty().Error()))
		default:
			slog.Warn("unknown upstream error with unknown Content-Type",
				slog.Int("status", res.StatusCode),
				slog.String("content-type", res.Header.Get("Content-Type")),
				slog.String("content-length", res.Header.Get("Content-Length")),
			)
		}
	}

	return mo.Ok[any](c.Stream(http.StatusOK, "audio/mp3", res.Body))
}
