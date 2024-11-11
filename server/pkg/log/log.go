package log

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strconv"

	slogMulti "github.com/samber/slog-multi"
)

// NewLogger returns a new logger with the given log level.
func init() {
	var level slog.Level = slog.LevelInfo
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		err := level.UnmarshalText([]byte(logLevel))
		if err != nil {
			panic(err)
		}
	}

	logger := slog.New(
		slogMulti.
			Pipe(
				// Add caller to the log record if the log level is error.
				slogMulti.NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
					if record.Level == slog.LevelError {
						record.AddAttrs(slog.String("caller", caller(5)()))
					}
					return next(ctx, record)
				}),
			).Handler(
			// Log to stderr.
			slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
			})),
	)

	slog.SetDefault(logger)
}

func caller(depth int) func() string {
	return func() string {
		_, file, line, _ := runtime.Caller(depth)
		return file + ":" + strconv.Itoa(line)
	}
}
