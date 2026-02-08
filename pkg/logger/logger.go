package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint" // Using tint for colored logs in Dev
	"github.com/mattn/go-isatty"
)

// Global logger instance
var Log *slog.Logger

// InitLogger initializes the global logger based on environment
// env: "dev" (text + color) or "prod" (json)
func InitLogger(env string) {
	var handler slog.Handler

	if env == "prod" {
		// Production: JSON format for ELK/Cloud logging
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Customize time format to RFC3339
				if a.Key == slog.TimeKey {
					return slog.Attr{Key: "timestamp", Value: slog.StringValue(a.Value.Time().Format(time.RFC3339))}
				}
				return a
			},
		})
	} else {
		// Development: Colored text output
		// Using lmittmann/tint for beautiful console logs
		w := os.Stderr
		handler = tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
			NoColor:    !isatty.IsTerminal(w.Fd()),
		})
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// WithContext adds trace_id from context to the logger
func WithContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return Log
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return Log.With("trace_id", traceID)
	}
	return Log
}

// Helper methods for quick access
func Info(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Info(msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Error(msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Debug(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).Warn(msg, args...)
}
