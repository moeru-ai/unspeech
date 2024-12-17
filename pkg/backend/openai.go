package backend

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/nekomeowww/fo"
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

	payload := fo.May(json.Marshal(values))

	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		"https://openai.com/v1/audio/speech",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	// TODO: Bearer Auth
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	defer res.Body.Close()

	// body, _ := io.ReadAll(res.Body)

	return mo.Ok[any](c.Stream(http.StatusOK, "audio/mp3", res.Body))
}
