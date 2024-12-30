package backend

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func openai(c echo.Context, options mo.Option[SpeechRequestOptions]) mo.Result[any] {
	values := OpenAISpeechRequestOptions{
		Model:          options.MustGet().Model,
		Input:          options.MustGet().Input,
		Voice:          options.MustGet().Voice,
		ResponseFormat: options.MustGet().ResponseFormat,
		Speed:          options.MustGet().Speed,
	}

	payload := lo.Must(json.Marshal(values))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://openai.com/v1/audio/speech",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	// Proxy the Authorization header
	req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
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
