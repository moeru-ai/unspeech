package logs

import (
	"log/slog"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

type CallerLike interface {
	GetFile() string
	GetLine() int64
	GetFunction() string
}

func Caller(c mo.Option[CallerLike]) []slog.Attr {
	if c.IsAbsent() || lo.IsNil(c) {
		return make([]slog.Attr, 0)
	}

	return []slog.Attr{
		slog.String("file", c.MustGet().GetFile()),
		slog.Int64("line", c.MustGet().GetLine()),
		slog.String("function", c.MustGet().GetFunction()),
	}
}
