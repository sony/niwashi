// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import "github.com/sony/niwashi/internal/merge"

// ServerCert controls how the WinRM server's TLS certificate is verified.
// It mirrors ssh.HostKey: trusting the party on the other end of the
// connection is a separate concern from Auth (the client's own credential
// material) and from Options (connection mechanics like the HTTP/HTTPS
// choice and timeouts).
type ServerCert struct {
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty"`
	CACertPath         string `json:"caCertPath,omitempty" yaml:"caCertPath,omitempty"`
}

func (s *ServerCert) Clone() *ServerCert {
	if s == nil {
		return nil
	}
	return &ServerCert{
		InsecureSkipVerify: s.InsecureSkipVerify,
		CACertPath:         s.CACertPath,
	}
}

func (s *ServerCert) Merge(other *ServerCert) *ServerCert {
	if other == nil {
		return s.Clone()
	}
	if s == nil {
		return other.Clone()
	}
	return &ServerCert{
		InsecureSkipVerify: s.InsecureSkipVerify || other.InsecureSkipVerify,
		CACertPath:         merge.String(s.CACertPath, other.CACertPath),
	}
}
