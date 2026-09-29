// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"context"
	"fmt"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/types"
)

type Prober struct {
	t  *Transport
	fs file.Opener
}

func NewProber(t *Transport, fs file.Opener) *Prober {
	return &Prober{t: t, fs: fs}
}

// GetSystemInfo queries the remote node's architecture and OS version over
// WinRM. Os is always platform.OsWindows since WinRM only talks to Windows
// hosts.
func (p *Prober) GetSystemInfo(ctx context.Context) (_ *platform.SystemInfo, err error) {
	client, err := NewClient(p.t, p.fs)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := client.Close(); e != nil {
			err = fmt.Errorf("%w (close: %s)", err, e)
		}
	}()

	cmd := `"{0}|{1}|{2}" -f $env:PROCESSOR_ARCHITECTURE, (Get-CimInstance Win32_OperatingSystem).Caption, (Get-CimInstance Win32_OperatingSystem).Version`
	var stdout, stderr strings.Builder
	if err := client.RunWithContext(ctx, cmd, &stdout, &stderr); err != nil {
		return nil, fmt.Errorf("failed to probe windows system info: %w (%s)", err, stderr.String())
	}

	fields := strings.SplitN(strings.TrimSpace(stdout.String()), "|", 3)
	if len(fields) != 3 {
		return nil, fmt.Errorf("unexpected system info output: %q", stdout.String())
	}

	return &platform.SystemInfo{
		Os:   platform.OsWindows,
		Arch: platform.NormalizeArch(fields[0]),
		Data: types.Dict{
			"caption": fields[1],
			"version": fields[2],
		},
	}, nil
}
