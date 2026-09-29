// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	"math/big"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/encoding/unicode"
)

// decodePowershellCommand reverses mwinrm.Powershell's encoding
// ("powershell.exe -EncodedCommand <base64 UTF16-LE>") so tests can assert
// on the original script text.
func decodePowershellCommand(t *testing.T, encoded string) string {
	t.Helper()
	const prefix = "powershell.exe -EncodedCommand "
	if !strings.HasPrefix(encoded, prefix) {
		t.Fatalf("command is not a %q invocation: %q", prefix, encoded)
	}
	utf16Bytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, prefix))
	if err != nil {
		t.Fatalf("failed to base64-decode command: %v", err)
	}
	decoded, err := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().Bytes(utf16Bytes)
	if err != nil {
		t.Fatalf("failed to UTF16-decode command: %v", err)
	}
	return string(decoded)
}

type fakeRunner struct {
	lastCmd  string
	cmds     []string
	stdout   string
	stderr   string
	exitCode int
	err      error
}

func (r *fakeRunner) RunWithContext(ctx context.Context, cmd string, stdout, stderr io.Writer) (int, error) {
	r.lastCmd = cmd
	r.cmds = append(r.cmds, cmd)
	if r.stdout != "" {
		_, _ = io.WriteString(stdout, r.stdout)
	}
	if r.stderr != "" {
		_, _ = io.WriteString(stderr, r.stderr)
	}
	return r.exitCode, r.err
}

func TestClient_RunWithContext(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		runErr   error
		wantErr  bool
	}{
		{name: "success", exitCode: 0, wantErr: false},
		{name: "non-zero exit code is an error", exitCode: 1, wantErr: true},
		{name: "transport error is propagated", exitCode: 0, runErr: io.ErrClosedPipe, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &fakeRunner{exitCode: tt.exitCode, err: tt.runErr}
			c := &client{winrm: r}

			var stdout, stderr strings.Builder
			err := c.RunWithContext(context.Background(), "whoami", &stdout, &stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_RunWithContext_WrapsAsPowershell(t *testing.T) {
	// Regression test: WinRM's default shell for a plain command line is
	// cmd.exe, not PowerShell. Every command this package builds is
	// PowerShell syntax, so it must be sent via "powershell.exe
	// -EncodedCommand ..." (mwinrm.Powershell), not as a raw command line --
	// otherwise cmd.exe rejects it (observed in practice as e.g. "'{0}' is
	// not recognized as an internal or external command").
	r := &fakeRunner{}
	c := &client{winrm: r}

	if err := c.RunWithContext(context.Background(), "Write-Host 'hi'", io.Discard, io.Discard); err != nil {
		t.Fatalf("RunWithContext() failed: %v", err)
	}

	decoded := decodePowershellCommand(t, r.lastCmd)
	if !strings.Contains(decoded, "Write-Host 'hi'") {
		t.Errorf("decoded command = %q, want to contain the original script", decoded)
	}
}

func TestClient_MkdirAll(t *testing.T) {
	r := &fakeRunner{}
	c := &client{winrm: r}

	if err := c.MkdirAll(context.Background(), `C:\Users\testuser\.niwashi`); err != nil {
		t.Fatalf("MkdirAll() failed: %v", err)
	}

	decoded := decodePowershellCommand(t, r.lastCmd)
	if !strings.Contains(decoded, `New-Item -ItemType Directory -Force -Path 'C:\Users\testuser\.niwashi'`) {
		t.Errorf("MkdirAll() command = %q, missing expected New-Item invocation", decoded)
	}
}

func TestClient_UploadContent(t *testing.T) {
	t.Run("uploads base64-encoded content", func(t *testing.T) {
		r := &fakeRunner{}
		c := &client{winrm: r}

		if err := c.UploadContent(context.Background(), `C:\file.txt`, []byte("hello")); err != nil {
			t.Fatalf("UploadContent() failed: %v", err)
		}

		decoded := decodePowershellCommand(t, r.lastCmd)

		// base64("hello") = aGVsbG8=
		if !strings.Contains(decoded, "aGVsbG8=") {
			t.Errorf("UploadContent() command = %q, missing expected base64 payload", decoded)
		}
		if !strings.Contains(decoded, `WriteAllBytes('C:\file.txt'`) {
			t.Errorf("UploadContent() command = %q, missing expected WriteAllBytes call", decoded)
		}
	})

	t.Run("rejects oversized content before running any command", func(t *testing.T) {
		r := &fakeRunner{}
		c := &client{winrm: r}

		err := c.UploadContent(context.Background(), `C:\big.bin`, make([]byte, maxUploadSize+1))
		if err == nil {
			t.Fatal("UploadContent() succeeded unexpectedly for oversized content")
		}
		if r.lastCmd != "" {
			t.Errorf("UploadContent() ran a command for oversized content: %q", r.lastCmd)
		}
	})
}

func TestClient_DownloadContent(t *testing.T) {
	r := &fakeRunner{stdout: "aGVsbG8=\r\n"}
	c := &client{winrm: r}

	got, err := c.DownloadContent(context.Background(), `C:\file.txt`)
	if err != nil {
		t.Fatalf("DownloadContent() failed: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("DownloadContent() = %q, want %q", got, "hello")
	}
}

func TestClient_ListDir(t *testing.T) {
	r := &fakeRunner{
		stdout: "1|sub\r\n0|sub\\inner.txt\r\n0|top.txt\r\n",
	}
	c := &client{winrm: r}

	entries, err := c.ListDir(context.Background(), `C:\remote`)
	if err != nil {
		t.Fatalf("ListDir() failed: %v", err)
	}

	want := []DirEntry{
		{IsDir: true, RelativePath: "sub"},
		{IsDir: false, RelativePath: `sub\inner.txt`},
		{IsDir: false, RelativePath: "top.txt"},
	}
	if len(entries) != len(want) {
		t.Fatalf("ListDir() returned %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, e := range entries {
		if e != want[i] {
			t.Errorf("ListDir() entry[%d] = %+v, want %+v", i, e, want[i])
		}
	}
}

// genSelfSignedCertPem generates a throwaway self-signed certificate for
// tests that need valid PEM bytes (e.g. ServerCert.CACertPath), mirroring
// internal/transport/ssh/auth_test.go's genPrivateKeyPem.
func genSelfSignedCertPem(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-ca"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create test certificate: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func newTestTransport(caCertPath string) *Transport {
	tr := &Transport{
		Address:    &Address{Host: "192.0.2.10", Port: 5986, User: "Administrator"},
		Auth:       &Auth{Method: AuthMethodNTLM, PasswordRef: &PasswordRef{FromEnv: "TEST_WINRM_CLIENT_PASSWORD"}},
		ServerCert: &ServerCert{CACertPath: caCertPath},
	}
	tr.SetDefaults()
	return tr
}

func TestNewClient_CACertPath(t *testing.T) {
	t.Setenv("TEST_WINRM_CLIENT_PASSWORD", "s3cr3t")

	t.Run("unset: no CA cert is read", func(t *testing.T) {
		fs := &mockFileOpener{files: map[string][]byte{}}
		if _, err := NewClient(newTestTransport(""), fs); err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}
	})

	t.Run("set but missing: wrapped read error", func(t *testing.T) {
		fs := &mockFileOpener{files: map[string][]byte{}}
		_, err := NewClient(newTestTransport("/ca/missing.pem"), fs)
		if err == nil {
			t.Fatal("NewClient() succeeded unexpectedly for a missing CA cert file")
		}
		if !strings.Contains(err.Error(), "failed to read winrm CA cert") {
			t.Errorf("NewClient() error = %v, want it to mention the CA cert read failure", err)
		}
	})

	t.Run("set with valid PEM: no error", func(t *testing.T) {
		fs := &mockFileOpener{files: map[string][]byte{"/ca/valid.pem": genSelfSignedCertPem(t)}}
		if _, err := NewClient(newTestTransport("/ca/valid.pem"), fs); err != nil {
			t.Fatalf("NewClient() failed with a valid CA cert: %v", err)
		}
	})

	t.Run("set with invalid content: error", func(t *testing.T) {
		fs := &mockFileOpener{files: map[string][]byte{"/ca/garbage.pem": []byte("not a cert")}}
		if _, err := NewClient(newTestTransport("/ca/garbage.pem"), fs); err == nil {
			t.Fatal("NewClient() succeeded unexpectedly for a non-PEM CA cert file")
		}
	})
}

func TestPsEscapeAndQuote(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantEscape string
		wantQuote  string
	}{
		{name: "no special chars", in: "hello", wantEscape: "hello", wantQuote: "'hello'"},
		{name: "single quote is doubled", in: "it's", wantEscape: "it''s", wantQuote: "'it''s'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := psEscape(tt.in); got != tt.wantEscape {
				t.Errorf("psEscape(%q) = %q, want %q", tt.in, got, tt.wantEscape)
			}
			if got := psQuote(tt.in); got != tt.wantQuote {
				t.Errorf("psQuote(%q) = %q, want %q", tt.in, got, tt.wantQuote)
			}
		})
	}
}
