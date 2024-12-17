package backend

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/mo"
)

func openai(options FullOptions) mo.Result[any] {
	values := Options{
		Model:          options.Model,
		Input:          options.Input,
		Voice:          options.Voice,
		ResponseFormat: options.ResponseFormat,
		Speed:          options.Speed,
	}

	body, err := json.Marshal(values)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	res, err := http.Post(
		"https://openai.com/v1/audio/speech",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return mo.Err[any](apierrors.NewErrBadRequest().WithCaller())
	}

	defer res.Body.Close()

	return mo.Ok[any](res)
}
