// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"context"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
)

func init() {
	// constructor
	transport.RegisterConstructor(
		TypeName,
		func() transport.Transport {
			return &Transport{}
		},
	)

	// operator
	transport.RegisterOperator(TypeName, func() transport.Operator {
		return &SshOperator{
			fs: &file.LocalFileSystem{},
		}
	})
}

type SshOperator struct {
	fs         file.FileSystem
	dryRunMode string
}

func (o *SshOperator) NewUpdater(operation string) (transport.Updater, error) {
	return NewKnownHostsUpdater(operation, o.fs), nil
}

func (o *SshOperator) NewProber(t transport.Transport) (transport.Prober, error) {
	if o.IsDryRun() {
		return NewSimProber(o.dryRunMode), nil
	} else {
		tp := t.(*Transport)
		return NewProber(tp, o.fs), nil
	}
}

func (o *SshOperator) NewRemoteSession(ctx context.Context, t transport.Transport) (transport.RemoteSession, error) {
	if o.IsDryRun() {
		return NewSimRemoteSession(o.dryRunMode), nil
	} else {
		tp := t.(*Transport)
		return NewRemoteSession(tp, o.fs), nil
	}
}

func (o *SshOperator) WithFileSystem(fs file.FileSystem) transport.Operator {
	return &SshOperator{
		fs:         fs,
		dryRunMode: o.dryRunMode,
	}
}

func (o *SshOperator) WithDryRunMode(mode string) transport.Operator {
	return &SshOperator{
		fs:         o.fs,
		dryRunMode: mode,
	}
}

func (o *SshOperator) IsDryRun() bool {
	return o.dryRunMode != "" && o.dryRunMode != "off"
}
