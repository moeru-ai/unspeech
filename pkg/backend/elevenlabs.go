package backend

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/utils/jsonpatch"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func elevenlabs(c echo.Context, options mo.Option[SpeechRequestOptions]) mo.Result[any] {
	reqURL := lo.Must(url.Parse("https://api.elevenlabs.io/v1/text-to-speech")).
		JoinPath(options.MustGet().Voice).
		String()

	// https://elevenlabs.io/docs/api-reference/text-to-speech/convert#request
	patchedPayload := jsonpatch.ApplyPatches(
		options.MustGet().body.OrElse(new(bytes.Buffer)).Bytes(),
		mo.Some(jsonpatch.ApplyOptions{AllowMissingPathOnRemove: true}),
		append(
			[]mo.Option[jsonpatch.JSONPatchOperationObject]{
				jsonpatch.NewRemove("/model"),
				jsonpatch.NewRemove("/voice"),
				jsonpatch.NewRemove("/input"),
				jsonpatch.NewAdd("/text", options.MustGet().Input),
				jsonpatch.NewAdd("/model_id", options.MustGet().Model),
			},
			lo.Map(
				lo.Entries(options.MustGet().ExtraBody),
				func(item lo.Entry[string, any], index int) mo.Option[jsonpatch.JSONPatchOperationObject] {
					return jsonpatch.NewAdd(strings.Join([]string{"/", item.Key}, ""), item.Value)
				})...,
		)...,
	)
	if patchedPayload.IsError() {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(patchedPayload.Error().Error()).WithCaller())
	}

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		reqURL,
		bytes.NewBuffer(patchedPayload.MustGet()),
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
				slog.String("content_type", res.Header.Get("Content-Type")),
				slog.String("content_length", res.Header.Get("Content-Length")),
			)
		}
	}

	return mo.Ok[any](c.Stream(http.StatusOK, "audio/mp3", res.Body))
}
