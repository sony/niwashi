// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh_test

import (
	"testing"

	"github.com/sony/niwashi/internal/transport/ssh"
)

func TestTransport_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		input   *ssh.Transport
		wantErr bool
	}{
		{
			name:    "empty config",
			input:   &ssh.Transport{},
			wantErr: true,
		},
		{
			name: "default config",
			input: func() *ssh.Transport {
				t := &ssh.Transport{}
				t.SetDefaults()
				return t
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.input.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}
