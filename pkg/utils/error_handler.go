package utils

import (
	"encoding/json"
	"io"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

var _ error = (*JSONResponseError)(nil)

type JSONResponseError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`

	bodyParsed map[string]any
}

func NewJSONResponseError(statusCode int, responseBody io.Reader) mo.Result[*JSONResponseError] {
	jsonData, err := io.ReadAll(responseBody)
	if err != nil {
		return mo.Err[*JSONResponseError](err)
	}

	resp := &JSONResponseError{
		StatusCode: statusCode,
	}

	err = json.Unmarshal(jsonData, &resp.bodyParsed)
	if err != nil {
		return mo.Err[*JSONResponseError](err)
	}

	errorMessage := GetByJSONPath[string](resp.bodyParsed, "{ .message }")
	errorStr := GetByJSONPath[string](resp.bodyParsed, "{ .error }")
	errorMap := GetByJSONPath[map[string]any](resp.bodyParsed, "{ .error }")
	errorStrFromErrorMap := GetByJSONPath[string](errorMap, "{ .message }")

	resp.Message = lo.Must(lo.Coalesce(errorMessage, errorStr, errorStrFromErrorMap, string(lo.Must(json.Marshal(resp.bodyParsed)))))

	return mo.Ok(resp)
}

func (r *JSONResponseError) Error() string {
	return r.Message
}

var _ error = (*TextResponseError)(nil)

type TextResponseError struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

func (r *TextResponseError) Error() string {
	return r.Body
}

func NewTextResponseError(statusCode int, responseBody io.Reader) mo.Result[*TextResponseError] {
	data, err := io.ReadAll(responseBody)
	if err != nil {
		return mo.Err[*TextResponseError](err)
	}

	return mo.Ok(&TextResponseError{
		StatusCode: statusCode,
		Body:       string(data),
	})
}
