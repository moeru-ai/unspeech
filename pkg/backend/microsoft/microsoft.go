package microsoft

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"strings"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/mo"
)

var (
	// NOTICE: Stupid Microsoft Speech / Azure AI Speech services returns error messages within the status text
	// therefore we attempt to obtain the error message from status text with enforced protocol to HTTP1.1.
	//
	// Try not to use this HTTP Client for the client implementation as much as possible, but for now it's the
	// only way to get the error message from the upstream.
	httpClient = &http.Client{
		Transport: &http.Transport{
			// According to documentation of https://pkg.go.dev/net/http#Transport.TLSNextProto:
			//
			// If TLSNextProto is not nil, HTTP/2 support is not enabled automatically.
			//
			// Source code: https://cs.opensource.google/go/go/+/master:src/net/http/transport.go;l=241-251
			// Thanks to https://go-review.googlesource.com/c/go/+/33094
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
	}
)

func handleResponseError(res *http.Response) mo.Result[any] {
	if res.Header.Get("Content-Length") == "" || res.Header.Get("Content-Length") == "0" {
		return mo.Err[any](apierrors.NewUpstreamError(res.StatusCode).WithDetail(res.Status))
	}

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

		return mo.Err[any](apierrors.NewUpstreamError(res.StatusCode).WithDetail(res.Status))
	}
}
