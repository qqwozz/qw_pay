package logger

import (
	"io"
	"log/slog"
)

var L *slog.Logger

func Setup(output io.Writer) {
	L = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(L)
}

func Info(msg string, args ...any) {
	L.Info(msg, args...)
}

func Error(msg string, args ...any) {
	L.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	L.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	L.Debug(msg, args...)
}
