package openai

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/backend/types"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	// Extract options safely once
	opt := options.MustGet()

	values := types.OpenAISpeechRequestOptions{
		Model:          opt.Model,
		Input:          opt.Input,
		Voice:          opt.Voice,
		ResponseFormat: opt.ResponseFormat,
		Speed:          opt.Speed,
	}

	payload := lo.Must(json.Marshal(values))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://api.openai.com/v1/audio/speech",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	req.Header.Set("Authorization", c.Request().Header.Get("Authorization"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return mo.Err[any](
			apierrors.NewErrBadGateway().
				WithDetail(err.Error()).
				WithError(err).
				WithCaller(),
		)
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= http.StatusBadRequest {
		ct := res.Header.Get("Content-Type")

		switch {
		case strings.HasPrefix(ct, "application/json"):
			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
			)
		case strings.HasPrefix(ct, "text/"):
			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail(utils.NewTextResponseError(res.StatusCode, res.Body).OrEmpty().Error()),
			)
		default:
			slog.Warn("unknown upstream error",
				slog.Int("status", res.StatusCode),
				slog.String("content_type", ct),
				slog.String("content_length", res.Header.Get("Content-Length")),
			)

			return mo.Err[any](
				apierrors.NewUpstreamError(res.StatusCode).
					WithDetail("unknown Content-Type: " + ct),
			)
		}
	}

	// Stream audio response with correct upstream content type
	return mo.Ok[any](c.Stream(http.StatusOK, res.Header.Get("Content-Type"), res.Body))
}
