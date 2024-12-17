package backend

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/mo"
)

func openai(c echo.Context, options FullOptions) mo.Result[any] {
	values := Options{
		Model:          options.Model,
		Input:          options.Input,
		Voice:          options.Voice,
		ResponseFormat: options.ResponseFormat,
		Speed:          options.Speed,
	}

	payload, err := json.Marshal(values)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	res, err := http.Post(
		"https://openai.com/v1/audio/speech",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	return mo.Ok[any](c.Blob(http.StatusOK, "audio/mp3", body))
}
