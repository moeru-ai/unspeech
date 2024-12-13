package logs

import (
	"log/slog"

	"github.com/samber/lo"
)

type CallerLike interface {
	GetFile() string
	GetLine() int64
	GetFunction() string
}

func Caller(c CallerLike) []slog.Attr {
	if lo.IsNil(c) {
		return make([]slog.Attr, 0)
	}

	return []slog.Attr{
		slog.String("file", c.GetFile()),
		slog.Int64("line", c.GetLine()),
		slog.String("function", c.GetFunction()),
	}
}
