// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"testing"

	"github.com/sony/niwashi/internal/state"
)

func TestState_Stage(t *testing.T) {
	t.Run("default stage refers to the fields of State", func(t *testing.T) {
		s := state.NewState()
		got, err := s.Stage(state.DefaultStageName)
		if err != nil {
			t.Fatalf("Stage() error = %v, want nil", err)
		}
		if got == nil {
			t.Fatalf("Stage() = nil, want non-nil")
		}
		if got.Inventory != s.Inventory || got.Infrastructure != s.Infrastructure || got.Runtime != s.Runtime {
			t.Errorf("Stage() must refer to the fields of State")
		}
	})

	t.Run("unsupported stage name", func(t *testing.T) {
		s := state.NewState()
		if got, err := s.Stage("bootstrap"); got != nil || err == nil {
			t.Errorf("Stage() = %+v, err = %v, want nil, want error", got, err)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var s *state.State
		if got, err := s.Stage(state.DefaultStageName); got != nil || err == nil {
			t.Errorf("Stage() = %+v, err = %v, want nil, want error", got, err)
		}
	})
}
