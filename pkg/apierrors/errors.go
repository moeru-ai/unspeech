package apierrors

import (
	"net/http"
)

func NewErrBadRequest() *Error {
	return NewError(http.StatusBadRequest, "BAD_REQUEST").
		WithTitle("Bad Request").
		WithDetail("The request was invalid or cannot be served")
}

func NewErrInvalidArgument() *Error {
	return NewError(http.StatusBadRequest, "INVALID_ARGUMENT").
		WithTitle("Invalid Argument").
		WithDetail("Invalid parameters, queries, body, or headers were sent, please check the request")
}

func NewErrRequiredFieldMissing() *Error {
	return NewError(http.StatusBadRequest, "REQUIRED_FIELD_MISSING").
		WithTitle("One or more required fields are missing").
		WithDetail("One or more required fields are missing, please verify and try again")
}

func NewErrUnauthorized() *Error {
	return NewError(http.StatusUnauthorized, "UNAUTHORIZED").
		WithTitle("Unauthorized").
		WithDetail("The requested resources require authentication")
}

func NewErrPermissionDenied() *Error {
	return NewError(http.StatusForbidden, "PERMISSION_DENIED").
		WithTitle("Permission Denied").
		WithDetail("You do not have permission to access the requested resources")
}

func NewErrForbidden() *Error {
	return NewError(http.StatusForbidden, "FORBIDDEN").
		WithTitle("Forbidden").
		WithDetail("You do not have permission to access the requested resources")
}

func NewErrNotFound() *Error {
	return NewError(http.StatusNotFound, "NOT_FOUND").
		WithTitle("Not Found").
		WithDetail("The requested resources were not found")
}

func NewErrPaymentRequired() *Error {
	return NewError(http.StatusPaymentRequired, "PAYMENT_REQUIRED").
		WithTitle("Payment Required").
		WithDetail("The requested resources require payment")
}

func NewErrQuotaExceeded() *Error {
	return NewError(http.StatusTooManyRequests, "QUOTA_EXCEEDED").
		WithTitle("Quota Exceeded").
		WithDetail("The request quota has been exceeded")
}

func NewErrInternal() *Error {
	return NewError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR").
		WithTitle("Internal Server Error").
		WithDetail("An internal server error occurred")
}

func NewErrBadGateway() *Error {
	return NewError(http.StatusBadGateway, "BAD_GATEWAY").
		WithTitle("Bad gateway").
		WithDetail("The server received an invalid response from an upstream server")
}

func NewErrUnavailable() *Error {
	return NewError(http.StatusServiceUnavailable, "UNAVAILABLE").
		WithTitle("Service Unavailable").
		WithDetail("The requested service is unavailable")
}

func NewUpstreamError(statusCode int) *Error {
	return NewError(statusCode, "UPSTREAM_ERROR").
		WithTitle("Upstream Error").
		WithReason("An error occurred while processing the request from the upstream service")
}
