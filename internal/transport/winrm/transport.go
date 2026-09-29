// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"fmt"

	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/transport"
)

const (
	TypeName = "winrm"
)

type Transport struct {
	Address    *Address    `json:"address,omitempty" yaml:"address,omitempty"`
	Auth       *Auth       `json:"auth,omitempty" yaml:"auth,omitempty"`
	Options    *Options    `json:"options,omitempty" yaml:"options,omitempty"`
	ServerCert *ServerCert `json:"serverCert,omitempty" yaml:"serverCert,omitempty"`
}

func (t *Transport) GetType() string {
	return TypeName
}

func (t *Transport) SetDefaults() {
	if t.Options == nil {
		t.Options = &Options{}
	}
	t.Options.SetDefaults()
	if t.Address == nil {
		t.Address = &Address{}
	}
	t.Address.SetDefaults(t.Options.UseHTTPS())
	if t.Auth == nil {
		t.Auth = &Auth{}
	}
	t.Auth.SetDefaults()
	if t.ServerCert == nil {
		t.ServerCert = &ServerCert{}
	}
}

func (t *Transport) Validate() error {
	if t.Address == nil {
		return fmt.Errorf("address is required for winrm connection")
	} else {
		if t.Address.Host == "" {
			return fmt.Errorf("address.host is required for winrm connection")
		}
		if t.Address.User == "" {
			return fmt.Errorf("address.user is required for winrm connection")
		}
	}
	if t.Auth == nil {
		return fmt.Errorf("auth is required for winrm connection")
	} else {
		switch t.Auth.Method {
		case AuthMethodNTLM, "":
			if t.Auth.PasswordRef == nil {
				return fmt.Errorf("auth.passwordRef is required for winrm connection with ntlm method")
			}
		default:
			return fmt.Errorf("unsupported auth method: %s", t.Auth.Method)
		}
	}
	return nil
}

func (t *Transport) Clone() transport.Transport {
	if t == nil {
		return nil
	}
	return &Transport{
		Address:    t.Address.Clone(),
		Auth:       t.Auth.Clone(),
		Options:    t.Options.Clone(),
		ServerCert: t.ServerCert.Clone(),
	}
}

func (t *Transport) Merge(o transport.Transport) (transport.Transport, error) {
	other, ok := o.(*Transport)
	if !ok {
		return nil, fmt.Errorf("cannot merge transport: incompatible types")
	}

	if other == nil {
		return t.Clone(), nil
	}
	if t == nil {
		return other.Clone(), nil
	}

	return &Transport{
		Address:    t.Address.Merge(other.Address),
		Auth:       t.Auth.Merge(other.Auth),
		Options:    t.Options.Merge(other.Options),
		ServerCert: t.ServerCert.Merge(other.ServerCert),
	}, nil
}

func (t *Transport) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	return fmt.Errorf("unsupported path: %s", path[0])
}

func (t *Transport) Remove(path []string) error {
	return fmt.Errorf("not implemented")
}

func (t *Transport) Update(o patch.Patchable) error {
	other, ok := o.(*Transport)
	if !ok {
		return fmt.Errorf("cannot update transport: incompatible types")
	}
	*t = *other
	return nil
}

func (t *Transport) Resolve(render func(string) (string, error)) error {
	if t.ServerCert != nil && t.ServerCert.CACertPath != "" {
		newPath, err := render(t.ServerCert.CACertPath)
		if err != nil {
			return fmt.Errorf("failed to render serverCert.caCertPath: %w", err)
		}
		t.ServerCert.CACertPath = newPath
	}
	return nil
}
