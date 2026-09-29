// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	Debug        = slog.Debug
	DebugContext = slog.DebugContext
	Info         = slog.Info
	InfoContext  = slog.InfoContext
	Warn         = slog.Warn
	WarnContext  = slog.WarnContext
	Error        = slog.Error
	ErrorContext = slog.ErrorContext
)

const (
	LevelTrace    = slog.LevelDebug - 1
	LevelDebug    = slog.LevelDebug
	LevelInfo     = slog.LevelInfo
	LevelProgress = slog.LevelInfo + 1
	LevelWarn     = slog.LevelWarn
	LevelError    = slog.LevelError
	LevelFatal    = slog.LevelError + 1
)

var tblLevel = map[string]slog.Level{
	"TRACE":    LevelTrace,
	"DEBUG":    LevelDebug,
	"INFO":     LevelInfo,
	"PROGRESS": LevelProgress,
	"WARN":     LevelWarn,
	"ERROR":    LevelError,
	"FATAL":    LevelFatal,
}

func Trace(msg string, args ...any) {
	slog.Log(context.Background(), LevelTrace, msg, args...)
}

func TraceContext(ctx context.Context, msg string, args ...any) {
	slog.Log(ctx, LevelTrace, msg, args...)
}
func Fatal(msg string, args ...any) {
	FatalContext(context.Background(), msg, args...)
}

func FatalContext(ctx context.Context, msg string, args ...any) {
	slog.Log(ctx, LevelFatal, msg, args...)
	panic("fatal error occurred, see logs for details")
}

func Progress(msg string, args ...any) {
	slog.Log(context.Background(), LevelProgress, msg, args...)
}

func ProgressContext(ctx context.Context, msg string, args ...any) {
	slog.Log(ctx, LevelProgress, msg, args...)
}

func GetLogLevel() slog.Level {
	level := os.Getenv("NWS_LOG_LEVEL")
	slogLevel, found := tblLevel[level]
	if !found {
		slogLevel = LevelInfo
	}
	return slogLevel
}

type Config struct {
	LogFile             string
	LogFormat           string
	BuildConsoleHandler func() (slog.Handler, error)
}

func InitLogger(cfg Config) error {

	handlers := []slog.Handler{}

	// init console handler
	handler, err := cfg.BuildConsoleHandler()
	if err != nil {
		return err
	}
	handlers = append(handlers, handler)

	if cfg.LogFile != "" {
		var writer io.Writer
		if cfg.LogFile == "-" {
			// stdout
			writer = os.Stdout
		} else {
			// file
			w, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			writer = w
		}

		handler, err := buildHandler(writer, cfg.LogFormat)
		if err != nil {
			return err
		}

		handlers = append(handlers, handler)
	}

	// todo: replace with slog.NewMultiHandler when it becomes available
	slog.SetDefault(slog.New(NewMultiHandler(handlers...)))

	Info("Log level", "level", GetLogLevel())

	return nil
}

func Finalize() {
	// Log file flush/close is delegated to the OS; nothing to do here
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		switch level {
		case LevelTrace:
			a.Value = slog.StringValue("TRACE")
		case LevelProgress:
			a.Value = slog.StringValue("PROGRESS")
		case LevelFatal:
			a.Value = slog.StringValue("FATAL")
		}
	}
	return a
}

func GetHandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:       GetLogLevel(),
		ReplaceAttr: replaceAttr,
	}
}

func buildHandler(w io.Writer, format string) (slog.Handler, error) {

	options := GetHandlerOptions()

	switch format {
	case "json":
		return slog.NewJSONHandler(w, options), nil
	case "plain":
		return slog.NewTextHandler(w, options), nil
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}
}
