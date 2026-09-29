// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package merge_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/merge"
)

func TestStringMap(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		a    map[string]string
		b    map[string]string
		want map[string]string
	}{
		{
			name: "merges two maps",
			a:    map[string]string{"key1": "value1", "key2": "value2"},
			b:    map[string]string{"key2": "new_value2", "key3": "value3"},
			want: map[string]string{"key1": "value1", "key2": "new_value2", "key3": "value3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := merge.StringMap(tt.a, tt.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
