package jsonapi

import "github.com/moeru-ai/unspeech/pkg/logs"

type Links struct {
	// a string whose value is a URI-reference [RFC3986 Section 4.1] pointing to the link’s target.
	Href string `json:"href,omitempty"`
	// a string indicating the link’s relation type.
	Rel *string `json:"rel,omitempty"`
	// a link to a description document (e.g. OpenAPI or JSON Schema) for the link target.
	Describedby *string `json:"describedby,omitempty"`
	// a string which serves as a label for the destination of a link
	// such that it can be used as a human-readable identifier (e.g., a menu entry).
	Title *string `json:"title,omitempty"`
	// a string indicating the media type of the link’s target.
	Type *string `json:"type,omitempty"`
	// a string or an array of strings indicating the language(s) of the link’s target.
	// An array of strings indicates that the link’s target is available in multiple languages.
	Hreflang *string `json:"hreflang,omitempty"`
	// a meta object containing non-standard meta-information about the link.
	Meta map[string]any `json:"meta,omitempty"`
}

type ErrorObjectSource struct {
	// a JSON Pointer [RFC6901] to the value in the request document that caused the error
	// [e.g. "/data" for a primary data object, or "/data/attributes/title" for a specific attribute].
	Pointer string `json:"pointer,omitempty"`
	// a string indicating which URI query parameter caused the error.
	Parameter string `json:"parameter,omitempty"`
	// a string indicating the name of a single request header which caused the error.
	Header string `json:"header,omitempty"`
}

// JSON API spec error object
// Documentation: https://jsonapi.org/format/#error-objects
type ErrorObject struct {
	// a unique identifier for this particular occurrence of the problem.
	ID string `json:"id,omitempty"`
	// a links object containing references to the source of the error.
	Links *Links `json:"links,omitempty"`
	// the HTTP status code applicable to this problem, expressed as a string value.
	Status int `json:"status,omitempty"`
	// the HTTP status code applicable to this problem, expressed as a string value.
	GrpcStatus uint64 `json:"grpc_status,omitempty"`
	// an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`
	// a short, human-readable summary of the problem
	Title string `json:"title,omitempty"`
	// a human-readable explanation specific to this occurrence of the problem. Like title.
	Detail string `json:"detail,omitempty"`
	// an object containing references to the source of the error.
	Source *ErrorObjectSource `json:"source,omitempty"`
	// a meta object containing non-standard meta-information about the error.
	Meta map[string]any `json:"meta,omitempty"`
}

var _ logs.CallerLike = (*ErrorCaller)(nil)

type ErrorCaller struct {
	File     string `json:"file,omitempty"`
	Line     int64  `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

func (e *ErrorCaller) GetFile() string {
	return e.File
}

func (e *ErrorCaller) GetLine() int64 {
	return e.Line
}

func (e *ErrorCaller) GetFunction() string {
	return e.Function
}
