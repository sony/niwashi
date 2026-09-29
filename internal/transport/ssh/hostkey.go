// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/patch"
)

const (
	HostKeyMethodKnownHostsFile = "knownHostsFile"
	HostKeyMethodDefault        = HostKeyMethodKnownHostsFile
)

type HostKey struct {
	Method         string `json:"method,omitempty" yaml:"method,omitempty"` // knownHostsFile
	KnownHostsPath string `json:"knownHostsPath,omitempty" yaml:"knownHostsPath,omitempty"`
}

func (hk *HostKey) SetDefaults() {
	if hk.Method == "" {
		hk.Method = HostKeyMethodDefault
	}
}

func (hk *HostKey) Clone() *HostKey {
	if hk == nil {
		return nil
	}
	return &HostKey{
		Method:         hk.Method,
		KnownHostsPath: hk.KnownHostsPath,
	}
}

func (hk *HostKey) Merge(other *HostKey) *HostKey {
	return &HostKey{
		Method:         merge.String(hk.Method, other.Method),
		KnownHostsPath: merge.String(hk.KnownHostsPath, other.KnownHostsPath),
	}
}

func (hk *HostKey) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "knownHostsPath":
		return patch.AddSimpleValue(
			path[1:],
			value,
			func(v string) {
				hk.KnownHostsPath = v
			},
		)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (hk *HostKey) Remove(path []string) error {
	return fmt.Errorf("not implemented")
}

func (hk *HostKey) Update(o patch.Patchable) error {
	other, ok := o.(*HostKey)
	if !ok {
		return fmt.Errorf("cannot update hostKey: incompatible types")
	}

	logger.Warn("Updating hostKey")
	*hk = *other
	return nil
}
