// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package logger_test

import (
	"log/slog"
	"testing"

	"github.com/sony/niwashi/internal/logger"
)

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		env  map[string]string
		want slog.Level
	}{
		{
			name: "default",
			env:  map[string]string{},
			want: slog.LevelInfo,
		},
		{
			name: "trace",
			env:  map[string]string{"NWS_LOG_LEVEL": "TRACE"},
			want: logger.LevelTrace,
		},
		{
			name: "debug",
			env:  map[string]string{"NWS_LOG_LEVEL": "DEBUG"},
			want: logger.LevelDebug,
		},
		{
			name: "wrong log level",
			env:  map[string]string{"NWS_LOG_LEVEL": "WRONG"},
			want: logger.LevelInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got := logger.GetLogLevel()
			if got != tt.want {
				t.Errorf("GetLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
