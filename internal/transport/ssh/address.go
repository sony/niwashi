// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"github.com/sony/niwashi/internal/merge"
)

type Address struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	User string `json:"user" yaml:"user"`
}

func (a *Address) SetDefaults() {
	if a.Port == 0 {
		a.Port = 22
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
