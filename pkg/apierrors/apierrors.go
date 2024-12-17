package apierrors

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/moeru-ai/unspeech/pkg/jsonapi"
	"github.com/moeru-ai/unspeech/pkg/logs"
	"github.com/samber/mo"
)

type Error struct {
	*jsonapi.ErrorObject

	caller   *jsonapi.ErrorCaller
	rawError error
}

func (e *Error) Error() string {
	return e.Detail
}

func (e *Error) AsResponse() *ErrResponse {
	return NewErrResponse().WithError(e)
}

func (e *Error) Caller() mo.Option[logs.CallerLike] {
	if e == nil || e.caller == nil {
		return mo.None[logs.CallerLike]()
	}

	return mo.Some(logs.CallerLike(e.caller))
}

func NewError[S ~int](status S, code string) *Error {
	return &Error{
		ErrorObject: &jsonapi.ErrorObject{
			ID:     code,
			Status: int(status),
			Code:   code,
		},
	}
}

func (e *Error) WithError(err error) *Error {
	e.rawError = err
	e.Detail = err.Error()

	return e
}

func (e *Error) WithCaller() *Error {
	pc, file, line, _ := runtime.Caller(1)

	e.caller = &jsonapi.ErrorCaller{
		Function: runtime.FuncForPC(pc).Name(),
		File:     file,
		Line:     int64(line),
	}

	return e
}

func (e *Error) WithTitle(title string) *Error {
	e.Title = title

	return e
}

func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail

	return e
}

func (e *Error) WithDetailf(format string, args ...any) *Error {
	e.Detail = fmt.Sprintf(format, args...)

	return e
}

func (e *Error) WithSourcePointer(pointer string) *Error {
	e.Source = mo.Some(jsonapi.ErrorObjectSource{
		Pointer: pointer,
	})

	return e
}

func (e *Error) WithSourceParameter(parameter string) *Error {
	e.Source = mo.Some(jsonapi.ErrorObjectSource{
		Parameter: parameter,
	})

	return e
}

func (e *Error) WithSourceHeader(header string) *Error {
	e.Source = mo.Some(jsonapi.ErrorObjectSource{
		Header: header,
	})

	return e
}

func (e *Error) WithReason(reason string) *Error {
	return e.WithMeta("reason", reason)
}

func (e *Error) WithMeta(key string, val any) *Error {
	if e.Meta.IsAbsent() {
		e.Meta = mo.Some(map[string]any{})
	}

	e.Meta.MustGet()[key] = val

	return e
}

type ErrResponse struct {
	jsonapi.Response
}

func NewErrResponseFromErrorObjects(errs ...*jsonapi.ErrorObject) *ErrResponse {
	resp := NewErrResponse()

	for _, err := range errs {
		resp = resp.WithError(&Error{
			ErrorObject: err,
		})
	}

	return resp
}

func NewErrResponseFromErrorObject(err *jsonapi.ErrorObject) *ErrResponse {
	return NewErrResponse().WithError(&Error{
		ErrorObject: err,
	})
}

func NewErrResponse() *ErrResponse {
	return &ErrResponse{
		Response: jsonapi.Response{
			Errors: make([]*jsonapi.ErrorObject, 0),
		},
	}
}

func (e *ErrResponse) WithError(err *Error) *ErrResponse {
	e.Errors = append(e.Errors, err.ErrorObject)

	return e
}

func (e *ErrResponse) HTTPStatus() int {
	if len(e.Errors) == 0 {
		return http.StatusOK
	}

	return e.Errors[0].Status
}
