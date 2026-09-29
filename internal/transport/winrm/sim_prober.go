// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"context"

	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/types"
)

// SimProber mirrors internal/transport/ssh's SimProber for --dry-run.
type SimProber struct {
	mode string
}

func NewSimProber(mode string) *SimProber {
	return &SimProber{mode: mode}
}

func (p *SimProber) GetSystemInfo(_ context.Context) (*platform.SystemInfo, error) {
	return &platform.SystemInfo{
		Os:   platform.OsWindows,
		Arch: platform.ArchAMD64,
		Data: types.Dict{
			"sim_mode": p.mode,
		},
	}, nil
}
