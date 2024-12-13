package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nekomeowww/fo"
	"github.com/samber/lo"
)

func ResponseLogV2() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			stop := time.Now()

			fields := []slog.Attr{
				slog.String("time", time.Now().Format(time.RFC3339)),
				slog.String("remote_ip", c.RealIP()),
				slog.String("host", req.Host),
				slog.String("uri", req.RequestURI),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("route", c.Path()),
				slog.String("protocol", req.Proto),
				slog.String("referer", req.Referer()),
				slog.String("user_agent", req.UserAgent()),
				slog.Int("status", res.Status),
				slog.Int64("latency", int64(stop.Sub(start))),
				slog.String("latency_human", stop.Sub(start).String()),
				slog.Int64("bytes_in", req.ContentLength),
				slog.Int64("bytes_out", res.Size),
				slog.Any("headers", req.Header),
				slog.Any("query", c.QueryParams()),
				slog.Any("form", fo.May(c.FormParams())),
				slog.Any("cookies", lo.Map(c.Cookies(), func(item *http.Cookie, index int) string {
					return item.String()
				})),
			}

			slog.Info("request", lo.ToAnySlice(fields)...)

			return err
		}
	}
}
