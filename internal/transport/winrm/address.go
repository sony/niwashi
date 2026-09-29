// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"github.com/sony/niwashi/internal/merge"
)

const (
	httpsPort = 5986
	httpPort  = 5985
)

type Address struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	User string `json:"user" yaml:"user"`
}

// SetDefaults fills in the port based on the scheme (https unless the
// caller has explicitly allowed insecure http, see Options.AllowInsecureHTTP).
func (a *Address) SetDefaults(useHTTPS bool) {
	if a.Port == 0 {
		if useHTTPS {
			a.Port = httpsPort
		} else {
			a.Port = httpPort
		}
	}
}

func (a *Address) Clone() *Address {
	if a == nil {
		return nil
	}
	return &Address{
		Host: a.Host,
		Port: a.Port,
		User: a.User,
	}
}

func (a *Address) Merge(other *Address) *Address {
	if other == nil {
		return a.Clone()
	}
	if a == nil {
		return other.Clone()
	}
	return &Address{
		Host: merge.String(a.Host, other.Host),
		Port: func() int {
			if other.Port != 0 {
				return other.Port
			}
			return a.Port
		}(),
		User: merge.String(a.User, other.User),
	}
}
