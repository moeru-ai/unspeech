package koemotion

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
	"github.com/moeru-ai/unspeech/pkg/utils/jsonpatch"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/vincent-petithory/dataurl"
)

func HandleSpeech(c echo.Context, options mo.Option[types.SpeechRequestOptions]) mo.Result[any] {
	// https://developers.rinna.co.jp/api-details#api=koemotion&operation=infer
	patchedPayload := jsonpatch.ApplyPatches(
		options.MustGet().AsBuffer().OrElse(new(bytes.Buffer)).Bytes(),
		mo.Some(jsonpatch.ApplyOptions{AllowMissingPathOnRemove: true}),
		append(
			[]mo.Option[jsonpatch.JSONPatchOperationObject]{
				jsonpatch.NewRemove("/model"),
				jsonpatch.NewRemove("/voice"),
				jsonpatch.NewRemove("/input"),
				jsonpatch.NewAdd("/text", options.MustGet().Input),
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
		"https://api.rinna.co.jp/koemotion/infer",
		bytes.NewBuffer(patchedPayload.MustGet()),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithCaller())
	}

	// Rewrite the Authorization header
	req.Header.Set("Ocp-Apim-Subscription-Key", strings.TrimPrefix(
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
				WithDetail(utils.NewJSONResponseError(res.StatusCode, res.Body).OrEmpty().Error()))
		case strings.HasPrefix(res.Header.Get("Content-Type"), "text/"):
			return mo.Err[any](apierrors.
				NewUpstreamError(res.StatusCode).
				WithDetail(utils.NewTextResponseError(res.StatusCode, res.Body).OrEmpty().Error()))
		default:
			slog.Warn("unknown upstream error with unknown Content-Type",
				slog.Int("status", res.StatusCode),
				slog.String("content_type", res.Header.Get("Content-Type")),
				slog.String("content_length", res.Header.Get("Content-Length")),
			)
		}
	}

	var resBody map[string]any

	err = json.NewDecoder(res.Body).Decode(&resBody)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	audioDataURLString := utils.GetByJSONPath[string](resBody, "{ .audio }")
	if audioDataURLString == "" {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail("upstream returned empty audio data URL").WithCaller())
	}

	audioDataURL, err := dataurl.DecodeString(audioDataURLString)
	if err != nil {
		return mo.Err[any](apierrors.NewErrInternal().WithDetail(err.Error()).WithError(err).WithCaller())
	}

	return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", audioDataURL.Data))
}
