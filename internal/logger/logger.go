package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	Log     *slog.Logger
	verbose = true
)

type noopHandler struct{}

func (h *noopHandler) Enabled(_ context.Context, _ slog.Level) bool { return false }
func (h *noopHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (h *noopHandler) WithAttrs(_ []slog.Attr) slog.Handler           { return h }
func (h *noopHandler) WithGroup(_ string) slog.Handler                { return h }

func Init(level slog.Level, v bool) {
	verbose = v

	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}
			return a
		},
	})

	if !verbose {
		handler = &noopHandler{}
	}

	Log = slog.New(handler)
}
