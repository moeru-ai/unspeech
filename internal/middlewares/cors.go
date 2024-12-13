package middlewares

import (
	"github.com/golang-module/carbon"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func CORS() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			return true, nil
		},
		AllowHeaders: []string{
			echo.HeaderAccept,
			echo.HeaderAcceptEncoding,
			// HeaderAllow is the name of the "Allow" header field used to list the set of methods
			// advertised as supported by the target resource. Returning an Allow header is mandatory
			// for status 405 (method not found) and useful for the OPTIONS method in responses.
			// See RFC 7231: https://datatracker.ietf.org/doc/html/rfc7231#section-7.4.1
			echo.HeaderAllow,
			echo.HeaderAuthorization,
			echo.HeaderContentDisposition,
			echo.HeaderContentEncoding,
			echo.HeaderContentLength,
			echo.HeaderContentType,
			echo.HeaderCookie,
			echo.HeaderSetCookie,
			echo.HeaderIfModifiedSince,
			echo.HeaderLastModified,
			echo.HeaderLocation,
			echo.HeaderRetryAfter,
			echo.HeaderUpgrade,
			echo.HeaderVary,
			echo.HeaderWWWAuthenticate,
			echo.HeaderXForwardedFor,
			echo.HeaderXForwardedProto,
			echo.HeaderXForwardedProtocol,
			echo.HeaderXForwardedSsl,
			echo.HeaderXUrlScheme,
			echo.HeaderXHTTPMethodOverride,
			echo.HeaderXRealIP,
			echo.HeaderXRequestID,
			echo.HeaderXCorrelationID,
			echo.HeaderXRequestedWith,
			echo.HeaderServer,
			echo.HeaderOrigin,
			echo.HeaderCacheControl,
			echo.HeaderConnection,
			// Access control
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAccessControlRequestHeaders,
			echo.HeaderAccessControlAllowOrigin,
			echo.HeaderAccessControlAllowMethods,
			echo.HeaderAccessControlAllowHeaders,
			echo.HeaderAccessControlAllowCredentials,
			echo.HeaderAccessControlExposeHeaders,
			echo.HeaderAccessControlMaxAge,
			// Security
			echo.HeaderStrictTransportSecurity,
			echo.HeaderXContentTypeOptions,
			echo.HeaderXXSSProtection,
			echo.HeaderXFrameOptions,
			echo.HeaderContentSecurityPolicy,
			echo.HeaderContentSecurityPolicyReportOnly,
			echo.HeaderXCSRFToken,
			echo.HeaderReferrerPolicy,
			// OpenAI & Vercel AI SDK Related
			"x-stainless-os",
			"x-stainless-lang",
			"x-stainless-package-version",
			"x-stainless-runtime",
			"x-stainless-arch",
			"x-stainless-runtime-version",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		MaxAge:       carbon.SecondsPerWeek,
		ExposeHeaders: []string{
			// OpenTelemetry
			"Traceparent",
			echo.HeaderContentType,
			echo.HeaderContentLength,
			echo.HeaderContentEncoding,
			echo.HeaderContentDisposition,
		},
	})
}
