// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import "testing"

// HasWorkingTask must only report tasks that have been dispatched and not yet
// completed. A registered producer whose tasks were never produced must not
// count: after an error no new tasks are produced, so counting such a
// producer would keep the run state from ever reaching its exit hooks.
func TestTaskGenerator_HasWorkingTask(t *testing.T) {
	tests := []struct {
		name      string
		producers map[string]*Producer
		want      bool
	}{
		{
			name:      "no producers",
			producers: map[string]*Producer{},
			want:      false,
		},
		{
			name: "registered producer with only unproduced tasks",
			producers: map[string]*Producer{
				"pending": {filteredTasks: []*FilteredTask{{}}},
			},
			want: false,
		},
		{
			name: "a producer with a dispatched task",
			producers: map[string]*Producer{
				"pending": {filteredTasks: []*FilteredTask{{}}},
				"running": {workingTasks: []*Task{{}}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &taskGenerator{producers: tt.producers}
			if got := g.HasWorkingTask(); got != tt.want {
				t.Errorf("HasWorkingTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
