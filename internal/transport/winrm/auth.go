// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"fmt"
	"os"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/merge"
)

// AuthMethodNTLM is the only supported auth method for v1 (design proposal
// 026, Non-Goals): Kerberos requires domain join, CredSSP isn't implemented
// by the underlying masterzen/winrm library, and Basic auth is disabled by
// default on WinRM servers and sends credentials with no encryption.
const (
	AuthMethodNTLM    = "ntlm"
	AuthMethodDefault = AuthMethodNTLM
)

type Auth struct {
	Method      string       `json:"method,omitempty" yaml:"method,omitempty"` // ntlm
	PasswordRef *PasswordRef `json:"passwordRef,omitempty" yaml:"passwordRef,omitempty"`

	// ClientCertPath and ClientKeyPath are reserved for the "certificate"
	// auth method (design proposal 026, later phase). Mirrors ssh.Auth's
	// PrivateKeyPath: the client's own credential material belongs in Auth,
	// as opposed to ServerCert.CACertPath, which verifies the server and
	// lives in its own field (mirroring ssh's separate HostKey field).
	// Unused for now -- "certificate" is not yet a supported Method value.
	ClientCertPath string `json:"clientCertPath,omitempty" yaml:"clientCertPath,omitempty"`
	ClientKeyPath  string `json:"clientKeyPath,omitempty" yaml:"clientKeyPath,omitempty"`
}

// PasswordRef is the source for the NTLM password. There is no plaintext
// alternative by design, so recipes/state never need the password written
// in plaintext; mirrors internal/transport/ssh's PassphraseRef.
type PasswordRef struct {
	FromEnv  string `json:"fromEnv,omitempty" yaml:"fromEnv,omitempty"`
	FromFile string `json:"fromFile,omitempty" yaml:"fromFile,omitempty"`
}

func (a *Auth) SetDefaults() {
	if a.Method == "" {
		a.Method = AuthMethodDefault
	}
}

func (a *Auth) Clone() *Auth {
	if a == nil {
		return nil
	}
	return &Auth{
		Method:         a.Method,
		PasswordRef:    a.PasswordRef.Clone(),
		ClientCertPath: a.ClientCertPath,
		ClientKeyPath:  a.ClientKeyPath,
	}
}

func (a *Auth) Merge(other *Auth) *Auth {
	if other == nil {
		return a.Clone()
	}
	if a == nil {
		return other.Clone()
	}
	return &Auth{
		Method:         merge.String(a.Method, other.Method),
		PasswordRef:    a.PasswordRef.Merge(other.PasswordRef),
		ClientCertPath: merge.String(a.ClientCertPath, other.ClientCertPath),
		ClientKeyPath:  merge.String(a.ClientKeyPath, other.ClientKeyPath),
	}
}

// GetPassword resolves the password via Auth.PasswordRef.
func (a *Auth) GetPassword(fs file.Opener) (string, error) {
	if a == nil {
		return "", fmt.Errorf("auth configuration is nil")
	}
	if a.PasswordRef == nil {
		return "", fmt.Errorf("auth.passwordRef is required")
	}
	return a.PasswordRef.GetPassword(fs)
}

func (p *PasswordRef) GetPassword(fs file.Opener) (string, error) {
	if p.FromEnv != "" {
		password := os.Getenv(p.FromEnv)
		if password == "" {
			return "", fmt.Errorf("environment variable %q is not set or empty", p.FromEnv)
		}
		return password, nil
	}
	if p.FromFile != "" {
		data, err := file.ReadFile(fs, p.FromFile)
		if err != nil {
			return "", fmt.Errorf("failed to read password file %q: %w", p.FromFile, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", fmt.Errorf("no password source specified")
}

func (p *PasswordRef) Clone() *PasswordRef {
	if p == nil {
		return nil
	}
	return &PasswordRef{
		FromEnv:  p.FromEnv,
		FromFile: p.FromFile,
	}
}

func (p *PasswordRef) Merge(other *PasswordRef) *PasswordRef {
	if other == nil {
		return p.Clone()
	}
	if p == nil {
		return other.Clone()
	}
	return &PasswordRef{
		FromEnv:  merge.String(p.FromEnv, other.FromEnv),
		FromFile: merge.String(p.FromFile, other.FromFile),
	}
}
