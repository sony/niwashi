// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/transport"
)

const (
	TypeName = "ssh"
)

type Transport struct {
	Address *Address `json:"address,omitempty" yaml:"address,omitempty"`
	Auth    *Auth    `json:"auth,omitempty" yaml:"auth,omitempty"`
	HostKey *HostKey `json:"hostKey,omitempty" yaml:"hostKey,omitempty"`
	Options *Options `json:"-" yaml:"-"`
}

func (t *Transport) GetType() string {
	return TypeName
}

func (t *Transport) SetDefaults() {
	if t.Address == nil {
		t.Address = &Address{}
	}
	t.Address.SetDefaults()
	if t.Auth == nil {
		t.Auth = &Auth{}
	}
	t.Auth.SetDefaults()
	if t.HostKey == nil {
		t.HostKey = &HostKey{}
	}
	t.HostKey.SetDefaults()
	if t.Options == nil {
		t.Options = &Options{}
	}
	t.Options.SetDefaults()
}

func (t *Transport) Validate() error {
	if t.Address == nil {
		return fmt.Errorf("address is required for ssh connection")
	} else {
		if t.Address.Host == "" {
			return fmt.Errorf("address.host is required for ssh connection")
		}
		if t.Address.User == "" {
			return fmt.Errorf("address.user is required for ssh connection")
		}
	}
	if t.Auth == nil {
		return fmt.Errorf("auth is required for ssh connection")
	} else {
		switch t.Auth.Method {
		case AuthMethodPrivateKey, "":
			if t.Auth.PrivateKeyPath == "" {
				return fmt.Errorf("auth.privateKeyPath is required for ssh connection with privateKey method")
			}
		default:
			return fmt.Errorf("unsupported auth method: %s", t.Auth.Method)
		}
	}
	if t.HostKey == nil {
		return fmt.Errorf("hostKey is required for ssh connection")
	} else {
		switch t.HostKey.Method {
		case HostKeyMethodKnownHostsFile, "":
			if t.HostKey.KnownHostsPath == "" {
				return fmt.Errorf("hostKey.knownHostsPath is required for ssh connection with knownHostsFile method")
			}
		default:
			return fmt.Errorf("unsupported hostKey method: %s", t.HostKey.Method)
		}
	}

	return nil
}

func (t *Transport) Clone() transport.Transport {
	if t == nil {
		return nil
	}
	clone := &Transport{
		Address: t.Address.Clone(),
		Auth:    t.Auth.Clone(),
		HostKey: t.HostKey.Clone(),
		Options: t.Options.Clone(),
	}
	return clone
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

	clone := &Transport{}
	var err error
	clone.Address = t.Address.Merge(other.Address)
	clone.Auth = t.Auth.Merge(other.Auth)
	clone.HostKey = t.HostKey.Merge(other.HostKey)
	clone.Options = t.Options.Merge(other.Options)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (t *Transport) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "hostKey":
		return patch.AddCompositeValue(
			path[1:],
			value,
			func() (*HostKey, bool) {
				return t.HostKey, true
			},
			func(hk *HostKey) {
				t.HostKey = hk
			},
		)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (t *Transport) Remove(path []string) error {
	return fmt.Errorf("not implemented")
}

func (t *Transport) Update(o patch.Patchable) error {
	other, ok := o.(*Transport)
	if !ok {
		return fmt.Errorf("cannot update transport: incompatible types")
	}

	logger.Warn("Updating transport")
	*t = *other
	return nil
}

func (t *Transport) Resolve(render func(string) (string, error)) error {
	if t.HostKey != nil {
		newHostPath, err := render(t.HostKey.KnownHostsPath)
		if err != nil {
			return fmt.Errorf("failed to render hostKey.knownHostsPath: %w", err)
		}
		t.HostKey.KnownHostsPath = newHostPath
	}

	return nil
}
