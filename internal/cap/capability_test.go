// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cap_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/types"
)

func TestCapability_Add(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		cap     *cap.Capability
		path    []string
		value   any
		expect  *cap.Capability
		wantErr bool
	}{
		{
			name: "invalid path",
			cap: &cap.Capability{
				Version: "0.0.1",
				Params:  make(map[string]any),
				Store:   make(map[string]any),
			},
			path:  []string{"invalid", "path"},
			value: "value",
			expect: &cap.Capability{
				Version: "0.0.1",
				Params:  make(map[string]any),
				Store:   make(map[string]any),
			},
			wantErr: true,
		},
		{
			name: "params: whole value replaced",
			cap: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
			path:  []string{"params"},
			value: map[string]any{"memory": "4096", "count": 3},
			expect: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "4096", "count": 3},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "params: invalid value type",
			cap: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
			},
			path:    []string{"params"},
			value:   "not-a-map",
			wantErr: true,
		},
		{
			// params must always be replaced as one coherent whole (see
			// Capability.ParamsHash): allowing a nested path to patch a single
			// field would let it drift out of sync with what was actually
			// resolved for the capability as a whole.
			name: "params: nested path is denied",
			cap: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
			},
			path:    []string{"params", "memory"},
			value:   "4096",
			wantErr: true,
		},
		{
			name: "version: whole value replaced, params and store kept",
			cap: &cap.Capability{
				Version: "1.0.0",
				Params:  types.Params{"memory": "1024"},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
			path:  []string{"version"},
			value: "1.1.0",
			expect: &cap.Capability{
				Version: "1.1.0",
				Params:  types.Params{"memory": "1024"},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "version: invalid value type",
			cap: &cap.Capability{
				Version: "1.0.0",
			},
			path:    []string{"version"},
			value:   110,
			wantErr: true,
		},
		{
			name: "version: nested path is denied",
			cap: &cap.Capability{
				Version: "1.0.0",
			},
			path:    []string{"version", "major"},
			value:   "1",
			wantErr: true,
		},
		{
			name: "store: whole value replaced",
			cap: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
			path:  []string{"store"},
			value: map[string]any{"ip": "10.0.0.2"},
			expect: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
				Store:   types.Dict{"ip": "10.0.0.2"},
			},
			wantErr: false,
		},
		{
			name: "store: invalid value type",
			cap: &cap.Capability{
				Version: "0.0.1",
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
			path:    []string{"store"},
			value:   123,
			wantErr: true,
		},
		{
			name: "store: nested path initializes a nil store",
			cap: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
				Store:   nil,
			},
			path:  []string{"store", "network", "ip"},
			value: "10.0.0.1",
			expect: &cap.Capability{
				Version: "0.0.1",
				Params:  types.Params{"memory": "1024"},
				Store:   types.Dict{"network": types.Dict{"ip": "10.0.0.1"}},
			},
			wantErr: false,
		},
		{
			name: "store: nested path preserves existing sibling data",
			cap: &cap.Capability{
				Version: "0.0.1",
				Store:   types.Dict{"existing": "value"},
			},
			path:  []string{"store", "network", "ip"},
			value: "10.0.0.1",
			expect: &cap.Capability{
				Version: "0.0.1",
				Store:   types.Dict{"existing": "value", "network": types.Dict{"ip": "10.0.0.1"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := tt.cap.Clone()

			gotErr := tt.cap.Add(tt.path, tt.value)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Add() failed: %v", gotErr)
				}
				if !reflect.DeepEqual(tt.cap, before) {
					t.Errorf("Add() mutated the capability despite returning an error: got %v, want unchanged %v", tt.cap, before)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Add() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(tt.cap, tt.expect) {
				t.Errorf("Add() = %v, want %v", tt.cap, tt.expect)
			}
		})
	}
}
