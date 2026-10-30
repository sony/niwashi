// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/workflow/planner"
)

// writeStateFiles writes each content to its own file and returns the paths.
func writeStateFiles(t *testing.T, contents ...string) []string {
	t.Helper()
	dir := t.TempDir()
	paths := make([]string, 0, len(contents))
	for i, c := range contents {
		p := filepath.Join(dir, "state"+string(rune('a'+i))+".yaml")
		if err := os.WriteFile(p, []byte(c), 0o600); err != nil {
			t.Fatalf("failed to write the state file: %v", err)
		}
		paths = append(paths, p)
	}
	return paths
}

func TestConfig_LoadTargetState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		wantErr string
	}{
		{
			name: "valid state",
			files: []string{`version: nws.state/v1
inventory:
  nodes:
    node1:
      capabilities:
        - test/node-capability@1.0.0
`},
		},
		{
			name: "reserved stage name",
			files: []string{`version: nws.state/v1
stages:
  main:
    inventory: {}
`},
			wantErr: `stage "main"`,
		},
		{
			name: "stage name reserved for the system",
			files: []string{`version: nws.state/v1
stages:
  _init:
    inventory: {}
`},
			wantErr: `stage "_init"`,
		},
		{
			name: "stage name not matching the pattern",
			files: []string{`version: nws.state/v1
stages:
  1bootstrap:
    inventory: {}
`},
			wantErr: `stage "1bootstrap"`,
		},
		{
			// the files are validated after merging: the node and the instance
			// bound to each other are in different files
			name: "binding split across files",
			files: []string{`version: nws.state/v1
inventory:
  nodes:
    node1:
      instanceRef: vm1:i1
`, `version: nws.state/v1
infrastructure:
  generators:
    vm1:
      provisioner: test/provisioner@1.0.0
      instances:
        i1:
          nodeRef: node1
`},
		},
		{
			name: "binding mismatch after merging",
			files: []string{`version: nws.state/v1
inventory:
  nodes:
    node1:
      instanceRef: vm1:i1
`, `version: nws.state/v1
infrastructure:
  generators:
    vm1:
      provisioner: test/provisioner@1.0.0
      instances:
        i1:
          nodeRef: node2
`},
			wantErr: "binding mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := planner.NewConfig(planner.WithTargetFiles(writeStateFiles(t, tt.files...)...))

			_, err := cfg.LoadTargetState()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("LoadTargetState() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("LoadTargetState() succeeded unexpectedly")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("LoadTargetState() error = %v, want %q in it", err, tt.wantErr)
			}
		})
	}
}
