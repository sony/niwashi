// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport/ssh"
	gossh "golang.org/x/crypto/ssh"
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

func (m *mockFile) Close() error {
	return nil
}

func genPrivateKeyPem(passphrase string) []byte {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	var pemBlock *pem.Block
	if passphrase != "" {
		pemBlock, _ = gossh.MarshalPrivateKeyWithPassphrase(key, "", []byte(passphrase))
	} else {
		pemBlock, _ = gossh.MarshalPrivateKey(key, "")
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	return pemBytes
}

func TestAuth_CreateSigner(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		fs      file.Opener
		auth    *ssh.Auth
		wantErr bool
	}{
		{
			name:    "nil auth",
			fs:      nil,
			auth:    nil,
			wantErr: true,
		},
		{
			name: "invalid private key path",
			fs: &mockFileOpener{
				files: map[string][]byte{},
			},
			auth: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/nonexistent/key",
			},
			wantErr: true,
		},
		{
			name: "invalid private key content",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/invalid/key": []byte(`-----BEGIN RSA PRIVATE KEY-----
INVALIDKEYCONTENT
-----END RSA PRIVATE KEY-----`),
				},
			},
			auth: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/invalid/key",
			},
			wantErr: true,
		},
		{
			name: "valid private key",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/valid/key": genPrivateKeyPem(""),
				},
			},
			auth: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/valid/key",
			},
			wantErr: false,
		},
		{
			name: "valid private key with passphrase",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/valid/key":       genPrivateKeyPem("testpassphrase"),
					"/passphrase/file": []byte("testpassphrase"),
				},
			},
			auth: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/valid/key",
				Passphrase: &ssh.PassphraseRef{
					FromFile: "/passphrase/file",
				},
			},
			wantErr: false,
		},
		{
			name: "valid private key with missing passphrase",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/valid/key":       genPrivateKeyPem("testpassphrase"),
					"/passphrase/file": []byte("wrongpassphrase"),
				},
			},
			auth: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/valid/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv: "TEST_PASSPHRASE_NOT_SET",
				},
			},
			wantErr: true,
		},
		{
			name: "valid private key with wrong passphrase",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/valid/key":       genPrivateKeyPem("testpassphrase"),
					"/passphrase/file": []byte("wrongpassphrase"),
				},
			},
			auth: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/valid/key",
				Passphrase: &ssh.PassphraseRef{
					FromFile: "/passphrase/file",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := tt.auth.CreateSigner(tt.fs)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateSigner() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateSigner() succeeded unexpectedly")
			}
		})
	}
}

func TestPassphraseRef_GetPassphrase(t *testing.T) {

	expectedPassphrase := "testpassphrase"

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fs      file.Opener
		p       *ssh.PassphraseRef
		want    string
		wantErr bool
	}{
		{
			name:    "empty PassphraseRef",
			fs:      nil,
			p:       &ssh.PassphraseRef{},
			want:    "",
			wantErr: true,
		},
		{
			name: "PassphraseRef with FromEnv",
			fs:   nil,
			p: &ssh.PassphraseRef{
				FromEnv: "TEST_PASSPHRASE",
			},
			want:    expectedPassphrase,
			wantErr: false,
		},
		{
			name: "PassphraseRef with FromEnv but env var is not set",
			fs:   nil,
			p: &ssh.PassphraseRef{
				FromEnv: "TEST_PASSPHRASE_NOT_SET",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "PassphraseRef with FromFile",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/passphrase/file": []byte(expectedPassphrase),
				},
			},
			p: &ssh.PassphraseRef{
				FromFile: "/passphrase/file",
			},
			want:    expectedPassphrase,
			wantErr: false,
		},
		{
			name: "PassphraseRef with FromFile and trime whitespace",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/passphrase/file": []byte(expectedPassphrase + "\n"),
				},
			},
			p: &ssh.PassphraseRef{
				FromFile: "/passphrase/file",
			},
			want:    expectedPassphrase,
			wantErr: false,
		},
		{
			name: "FromEnv and FromFile, FromEnv takes precedence",
			fs: &mockFileOpener{
				files: map[string][]byte{
					"/passphrase/file": []byte(expectedPassphrase + "fromfile"),
				},
			},
			p: &ssh.PassphraseRef{
				FromEnv:  "TEST_PASSPHRASE", // should take precedence over FromFile
				FromFile: "/passphrase/file",
			},
			want:    expectedPassphrase,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_PASSPHRASE", expectedPassphrase)

			got, gotErr := tt.p.GetPassphrase(tt.fs)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPassphrase() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPassphrase() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetPassphrase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_SetDefaults(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *ssh.Auth
	}{
		{
			name: "default Auth",
			want: &ssh.Auth{
				Method: ssh.AuthMethodPrivateKey,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a ssh.Auth
			a.SetDefaults()
			if !reflect.DeepEqual(&a, tt.want) {
				t.Errorf("SetDefaults() = %v, want %v", &a, tt.want)
			}
		})
	}
}

func TestAuth_Clone(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		arg  *ssh.Auth
		want *ssh.Auth
	}{
		{
			name: "nil Auth",
			arg:  nil,
			want: nil,
		},
		{
			name: "normal 1",
			arg:  &ssh.Auth{},
			want: &ssh.Auth{},
		},
		{
			name: "normal 2",
			arg: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv:  "TEST_PASSPHRASE",
					FromFile: "/path/to/passphrase",
				},
			},
			want: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv:  "TEST_PASSPHRASE",
					FromFile: "/path/to/passphrase",
				},
			},
		},
		{
			name: "nil Passphrase",
			arg: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase:     nil,
			},
			want: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase:     nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.Clone()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_Merge(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		arg   *ssh.Auth
		other *ssh.Auth
		want  *ssh.Auth
	}{
		{
			name:  "nil Auth",
			arg:   nil,
			other: nil,
			want:  nil,
		},
		{
			name:  "nil arg, non-nil other",
			arg:   nil,
			other: &ssh.Auth{},
			want:  &ssh.Auth{},
		},
		{
			name:  "non-nil arg, nil other",
			arg:   &ssh.Auth{},
			other: nil,
			want:  &ssh.Auth{},
		},
		{
			name: "overwrite",
			arg:  &ssh.Auth{},
			other: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv:  "TEST_PASSPHRASE",
					FromFile: "/path/to/passphrase",
				},
			},
			want: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv:  "TEST_PASSPHRASE",
					FromFile: "/path/to/passphrase",
				},
			},
		},
		{
			name: "overwrite2",
			arg: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv:  "TEST_PASSPHRASE",
					FromFile: "/path/to/passphrase",
				},
			},
			other: &ssh.Auth{},
			want: &ssh.Auth{
				Method:         ssh.AuthMethodPrivateKey,
				PrivateKeyPath: "/path/to/key",
				Passphrase: &ssh.PassphraseRef{
					FromEnv:  "TEST_PASSPHRASE",
					FromFile: "/path/to/passphrase",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.Merge(tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}
