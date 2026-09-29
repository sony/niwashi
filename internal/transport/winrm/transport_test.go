// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm_test

import (
	"testing"

	"github.com/sony/niwashi/internal/transport/winrm"
)

func TestTransport_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *winrm.Transport
		wantErr bool
	}{
		{
			name:    "empty config",
			input:   &winrm.Transport{},
			wantErr: true,
		},
		{
			name: "defaults without passwordRef",
			input: func() *winrm.Transport {
				tp := &winrm.Transport{
					Address: &winrm.Address{Host: "node1", User: "Administrator"},
				}
				tp.SetDefaults()
				return tp
			}(),
			wantErr: true,
		},
		{
			name: "valid ntlm config with passwordRef",
			input: func() *winrm.Transport {
				tp := &winrm.Transport{
					Address: &winrm.Address{Host: "node1", User: "Administrator"},
					Auth:    &winrm.Auth{PasswordRef: &winrm.PasswordRef{FromEnv: "WINRM_PASSWORD"}},
				}
				tp.SetDefaults()
				return tp
			}(),
			wantErr: false,
		},
		{
			name: "unsupported auth method",
			input: &winrm.Transport{
				Address: &winrm.Address{Host: "node1", User: "Administrator"},
				Auth:    &winrm.Auth{Method: "kerberos", PasswordRef: &winrm.PasswordRef{FromEnv: "WINRM_PASSWORD"}},
			},
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

func TestTransport_SetDefaults_Port(t *testing.T) {
	tests := []struct {
		name     string
		options  *winrm.Options
		wantPort int
	}{
		{name: "https by default", options: nil, wantPort: 5986},
		{name: "insecure http opt-in", options: &winrm.Options{AllowInsecureHTTP: true}, wantPort: 5985},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &winrm.Transport{
				Address: &winrm.Address{Host: "node1", User: "Administrator"},
				Options: tt.options,
			}
			tp.SetDefaults()
			if tp.Address.Port != tt.wantPort {
				t.Errorf("Address.Port = %d, want %d", tp.Address.Port, tt.wantPort)
			}
		})
	}
}

func TestTransport_Clone(t *testing.T) {
	tp := &winrm.Transport{
		Address: &winrm.Address{Host: "node1", Port: 5986, User: "Administrator"},
		Auth:    &winrm.Auth{Method: winrm.AuthMethodNTLM, PasswordRef: &winrm.PasswordRef{FromEnv: "WINRM_PASSWORD"}},
		Options: &winrm.Options{ConnectTimeoutSec: 30},
	}

	clone := tp.Clone().(*winrm.Transport)
	clone.Address.Host = "changed"

	if tp.Address.Host == "changed" {
		t.Error("Clone() did not deep-copy Address")
	}
}
