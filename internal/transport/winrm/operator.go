// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"context"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/workspace"
)

func init() {
	transport.RegisterConstructor(
		TypeName,
		func() transport.Transport {
			return &Transport{}
		},
	)

	transport.RegisterOperator(TypeName, func() transport.Operator {
		return &Operator{
			fs: &file.LocalFileSystem{},
		}
	})
}

type Operator struct {
	fs         file.FileSystem
	dryRunMode string
}

// NewUpdater returns a no-op updater: unlike SSH's known_hosts management,
// WinRM has no equivalent host-trust file to keep in sync with the
// workspace.
func (o *Operator) NewUpdater(operation string) (transport.Updater, error) {
	return &noopUpdater{}, nil
}

func (o *Operator) NewProber(t transport.Transport) (transport.Prober, error) {
	if o.IsDryRun() {
		return NewSimProber(o.dryRunMode), nil
	}
	tp := t.(*Transport)
	return NewProber(tp, o.fs), nil
}

func (o *Operator) NewRemoteSession(ctx context.Context, t transport.Transport) (transport.RemoteSession, error) {
	if o.IsDryRun() {
		return NewSimRemoteSession(o.dryRunMode), nil
	}
	return NewRemoteSession(t, o.fs), nil
}

func (o *Operator) WithFileSystem(fs file.FileSystem) transport.Operator {
	return &Operator{
		fs:         fs,
		dryRunMode: o.dryRunMode,
	}
}

func (o *Operator) WithDryRunMode(mode string) transport.Operator {
	return &Operator{
		fs:         o.fs,
		dryRunMode: mode,
	}
}

func (o *Operator) IsDryRun() bool {
	return o.dryRunMode != "" && o.dryRunMode != "off"
}

type noopUpdater struct{}

func (u *noopUpdater) Update(t transport.Transport, ws *workspace.RunWorkspace) ([]transport.Patch, error) {
	return nil, nil
}
