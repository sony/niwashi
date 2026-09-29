// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sony/niwashi/internal/logger"
)

func initPlanLogger(cfg logger.Config) {
	cfg.BuildConsoleHandler = func() (slog.Handler, error) {
		return &PlanLogHandler{}, nil
	}

	if err := logger.InitLogger(cfg); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

func finalizePlanLogger() {
	logger.Finalize()
}

type PlanLogHandler struct{}

func (m *PlanLogHandler) Handle(ctx context.Context, r slog.Record) error {

	if r.Level != logger.LevelProgress {
		return nil
	}

	switch r.Message {
	case "PlanStart":
		printPlanStart(&r)
	case "PlanComplete":
		printPlanComplete(&r)
	}

	return nil
}

func (m *PlanLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PlanLogHandler{}
}

func (m *PlanLogHandler) WithGroup(name string) slog.Handler {
	return &PlanLogHandler{}
}

func (m *PlanLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level == logger.LevelProgress
}

func printPlanStart(r *slog.Record) {
	fmt.Println("Plan started:")

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "targets":
			fmt.Printf("- Targets: %s\n", a.Value.String())
		case "profile":
			fmt.Printf("- Profile: %s\n", a.Value.String())
		case "destroy":
			fmt.Printf("- Destroy: %t\n", a.Value.Bool())
		case "prune":
			fmt.Printf("- Prune: %t\n", a.Value.Bool())
		}
		return true
	})
}

func printPlanComplete(r *slog.Record) {
	fmt.Println("Plan complete:")
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "out":
			fmt.Printf("- Output: %s\n", a.Value.String())
		case "error":
			e := a.Value.Any()
			if e != nil {
				fmt.Printf("- Error: %v\n", e.(error))
			}
		}
		return true
	})
}
