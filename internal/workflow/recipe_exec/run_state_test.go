// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/sony/niwashi/internal/workflow"
)

// saveStateContext implements only SaveState, which is the single
// ApplyContext method runState.OnExit uses. Calling any other method panics.
type saveStateContext struct {
	workflow.ApplyContext
	saveCalls int
	saveErr   error
}

func (c *saveStateContext) SaveState() error {
	c.saveCalls++
	return c.saveErr
}

func TestRunState_OnExit(t *testing.T) {
	const noHookError = -1

	tests := []struct {
		name string
		// error already held by the generator when the run state exits
		// (e.g. a task failed)
		errBeforeExit error
		saveErr       error
		// index of the success hook that holds an error, or noHookError
		successHookErrAt int
		wantCalls        []string
		wantErr          bool
	}{
		{
			name:             "no error: only success hooks run, in registration order",
			successHookErrAt: noHookError,
			wantCalls:        []string{"success-0", "success-1"},
			wantErr:          false,
		},
		{
			name:             "error before exit: only failure hooks run, in registration order",
			errBeforeExit:    errors.New("task failed"),
			successHookErrAt: noHookError,
			wantCalls:        []string{"failure-0", "failure-1"},
			wantErr:          true,
		},
		{
			// Success or failure is decided once when exit starts, so an
			// error raised by a success hook neither skips the remaining
			// success hooks nor switches to the failure hooks.
			name:             "error raised by a success hook: remaining success hooks still run",
			successHookErrAt: 0,
			wantCalls:        []string{"success-0", "success-1"},
			wantErr:          true,
		},
		{
			name:             "SaveState error is held on the generator",
			saveErr:          errors.New("save failed"),
			successHookErrAt: noHookError,
			wantCalls:        []string{"success-0", "success-1"},
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &saveStateContext{saveErr: tt.saveErr}
			g := &taskGenerator{ApplyContext: ctx}
			g.HoldError(tt.errBeforeExit)
			s := &runState{g: g}

			var calls []string
			for i := range 2 {
				s.AddSuccessFunc(func() {
					calls = append(calls, "success-"+strconv.Itoa(i))
					if i == tt.successHookErrAt {
						g.HoldError(errors.New("hook failed"))
					}
				})
				s.AddFailureFunc(func() {
					calls = append(calls, "failure-"+strconv.Itoa(i))
				})
			}

			s.OnExit()

			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("hooks called = %v, want %v", calls, tt.wantCalls)
			}
			if ctx.saveCalls != 1 {
				t.Errorf("SaveState called %d times, want 1", ctx.saveCalls)
			}
			if g.HasError() != tt.wantErr {
				t.Errorf("generator HasError() = %v, want %v (error: %v)", g.HasError(), tt.wantErr, g.GetError())
			}
		})
	}
}
