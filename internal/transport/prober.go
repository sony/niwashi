// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import (
	"context"

	"github.com/sony/niwashi/internal/platform"
)

type Prober interface {
	GetSystemInfo(ctx context.Context) (*platform.SystemInfo, error)
}

type ConstantProber struct {
	info *platform.SystemInfo
}

func NewConstantProber(info *platform.SystemInfo) *ConstantProber {
	return &ConstantProber{
		info: info.Clone(),
	}
}

func (p *ConstantProber) GetSystemInfo(_ context.Context) (*platform.SystemInfo, error) {
	return p.info, nil
}
