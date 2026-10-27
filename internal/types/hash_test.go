// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types_test

import (
	"testing"

	"github.com/sony/niwashi/internal/types"
)

func TestMakeHash(t *testing.T) {
	// the SHA-256 of the JSON form: printf '"PEM"' | sha256sum
	const hashPEM = "495d8f9c23c6e1682bfbd1e3570c6d8f85d799f00d4c43660e73921a86f7cf21"
	if got := types.MakeHash("PEM"); got != hashPEM {
		t.Errorf("MakeHash(%q) = %q, want %q", "PEM", got, hashPEM)
	}

	tests := []struct {
		name  string
		a, b  any
		equal bool
	}{
		{
			// a value read back from state.json has float64 numbers
			name:  "int and float64 of the same number",
			a:     map[string]any{"port": 3128},
			b:     map[string]any{"port": float64(3128)},
			equal: true,
		},
		{
			name:  "map with a different key order",
			a:     map[string]any{"host": "gateway", "port": 3128},
			b:     map[string]any{"port": 3128, "host": "gateway"},
			equal: true,
		},
		{
			name:  "string and number",
			a:     "1",
			b:     1,
			equal: false,
		},
		{
			name:  "different values",
			a:     map[string]any{"host": "gateway"},
			b:     map[string]any{"host": "resolver"},
			equal: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := types.MakeHash(tt.a) == types.MakeHash(tt.b); got != tt.equal {
				t.Errorf("MakeHash(%v) == MakeHash(%v) is %v, want %v", tt.a, tt.b, got, tt.equal)
			}
		})
	}

	t.Run("Dict.Hash uses the same hash", func(t *testing.T) {
		d := types.Dict{"host": "gateway", "port": 3128}
		if got, want := d.Hash(), types.MakeHash(map[string]any(d)); got != want {
			t.Errorf("Dict.Hash() = %q, want %q", got, want)
		}
	})
}
