// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

type mockFileOpener struct {
	files map[string][]byte
}

func (m *mockFileOpener) Open(path string) (io.ReadSeekCloser, error) {
	if content, ok := m.files[path]; ok {
		return &mockFile{Reader: bytes.NewReader(content)}, nil
	}
	return nil, fmt.Errorf("mockFileOpener: file not found: %s", path)
}

type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error { return nil }

func TestAuth_GetPassword(t *testing.T) {
	tests := []struct {
		name    string
		fs      *mockFileOpener
		auth    *Auth
		want    string
		wantErr bool
	}{
		{
			name:    "nil auth",
			auth:    nil,
			wantErr: true,
		},
		{
			name: "passwordRef with FromEnv",
			auth: &Auth{
				PasswordRef: &PasswordRef{FromEnv: "TEST_WINRM_PASSWORD"},
			},
			want: "s3cr3t",
		},
		{
			name: "passwordRef with FromFile",
			fs: &mockFileOpener{
				files: map[string][]byte{"/password/file": []byte("s3cr3t\n")},
			},
			auth: &Auth{
				PasswordRef: &PasswordRef{FromFile: "/password/file"},
			},
			want: "s3cr3t",
		},
		{
			name:    "no password or passwordRef",
			auth:    &Auth{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_WINRM_PASSWORD", "s3cr3t")

			var fs *mockFileOpener
			if tt.fs != nil {
				fs = tt.fs
			}
			got, gotErr := tt.auth.GetPassword(fs)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPassword() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPassword() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetPassword() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAuth_Clone(t *testing.T) {
	tests := []struct {
		name string
		arg  *Auth
		want *Auth
	}{
		{name: "nil Auth", arg: nil, want: nil},
		{
			name: "with passwordRef",
			arg: &Auth{
				Method:      AuthMethodNTLM,
				PasswordRef: &PasswordRef{FromEnv: "TEST_WINRM_PASSWORD", FromFile: "/path"},
			},
			want: &Auth{
				Method:      AuthMethodNTLM,
				PasswordRef: &PasswordRef{FromEnv: "TEST_WINRM_PASSWORD", FromFile: "/path"},
			},
		},
		{
			name: "with client cert/key paths",
			arg: &Auth{
				ClientCertPath: "/cert.pem",
				ClientKeyPath:  "/key.pem",
			},
			want: &Auth{
				ClientCertPath: "/cert.pem",
				ClientKeyPath:  "/key.pem",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.Clone()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAuth_Merge(t *testing.T) {
	tests := []struct {
		name  string
		base  *Auth
		other *Auth
		want  *Auth
	}{
		{
			name:  "other's client cert/key paths take precedence",
			base:  &Auth{ClientCertPath: "/base-cert.pem", ClientKeyPath: "/base-key.pem"},
			other: &Auth{ClientCertPath: "/other-cert.pem", ClientKeyPath: "/other-key.pem"},
			want:  &Auth{ClientCertPath: "/other-cert.pem", ClientKeyPath: "/other-key.pem"},
		},
		{
			name:  "empty other keeps base's client cert/key paths",
			base:  &Auth{ClientCertPath: "/base-cert.pem", ClientKeyPath: "/base-key.pem"},
			other: &Auth{},
			want:  &Auth{ClientCertPath: "/base-cert.pem", ClientKeyPath: "/base-key.pem"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.base.Merge(tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
