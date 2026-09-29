// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"context"

	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/types"
)

type SimProber struct {
	mode string
}

func NewSimProber(mode string) *SimProber {
	return &SimProber{mode: mode}
}

func (p *SimProber) GetSystemInfo(_ context.Context) (*platform.SystemInfo, error) {
	// return dummy system information
	return &platform.SystemInfo{
		Os:   platform.OsLinux,
		Arch: platform.ArchAMD64,
		Data: types.Dict{
			"sim_mode": p.mode,
		},
	}, nil
}
