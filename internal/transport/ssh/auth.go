// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/merge"
	"golang.org/x/crypto/ssh"
)

const (
	AuthMethodPrivateKey = "privateKey"
	AuthMethodDefault    = AuthMethodPrivateKey
)

type Auth struct {
	Method         string         `json:"method,omitempty" yaml:"method,omitempty"` // privateKey
	PrivateKeyPath string         `json:"privateKeyPath,omitempty" yaml:"privateKeyPath,omitempty"`
	Passphrase     *PassphraseRef `json:"passphraseRef,omitempty" yaml:"passphraseRef,omitempty"`
}

type PassphraseRef struct {
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
		PrivateKeyPath: a.PrivateKeyPath,
		Passphrase:     a.Passphrase.Clone(),
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
		PrivateKeyPath: merge.String(a.PrivateKeyPath, other.PrivateKeyPath),
		Passphrase:     a.Passphrase.Merge(other.Passphrase),
	}
}

func (a *Auth) CreateSigner(fs file.Opener) (ssh.Signer, error) {
	if a == nil {
		return nil, fmt.Errorf("auth configuration is nil")
	}

	keyBytes, err := file.ReadFile(fs, a.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key %q: %w", a.PrivateKeyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err == nil {
		return signer, nil
	}

	var perr *ssh.PassphraseMissingError
	if !errors.As(err, &perr) {
		return nil, fmt.Errorf("failed to parse private key %q: %w", a.PrivateKeyPath, err)
	}
	if a.Passphrase == nil {
		return nil, fmt.Errorf("private key %q is passphrase protected, but no passphrase provided", a.PrivateKeyPath)
	}

	passphrase, err := a.Passphrase.GetPassphrase(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to get passphrase for private key %q: %w", a.PrivateKeyPath, err)
	}

	signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key %q with passphrase: %w", a.PrivateKeyPath, err)
	}

	return signer, nil
}

func (p *PassphraseRef) GetPassphrase(fs file.Opener) (string, error) {
	if p.FromEnv != "" {
		passphrase := os.Getenv(p.FromEnv)
		if passphrase == "" {
			return "", fmt.Errorf("environment variable %q is not set or empty", p.FromEnv)
		}
		return passphrase, nil
	}
	if p.FromFile != "" {
		data, err := file.ReadFile(fs, p.FromFile)
		if err != nil {
			return "", fmt.Errorf("failed to read passphrase file %q: %w", p.FromFile, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", fmt.Errorf("no passphrase source specified")
}

func (r *PassphraseRef) Clone() *PassphraseRef {
	if r == nil {
		return nil
	}
	return &PassphraseRef{
		FromEnv:  r.FromEnv,
		FromFile: r.FromFile,
	}
}

func (r *PassphraseRef) Merge(other *PassphraseRef) *PassphraseRef {
	if other == nil {
		return r.Clone()
	}
	if r == nil {
		return other.Clone()
	}
	return &PassphraseRef{
		FromEnv:  merge.String(r.FromEnv, other.FromEnv),
		FromFile: merge.String(r.FromFile, other.FromFile),
	}
}
