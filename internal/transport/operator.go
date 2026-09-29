// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import (
	"context"
	"fmt"

	"github.com/sony/niwashi/internal/file"
)

const (
	OperationConstruct = "construct"
	OperationDestruct  = "destruct"
)

type Operator interface {
	NewRemoteSession(ctx context.Context, t Transport) (RemoteSession, error)
	NewUpdater(operation string) (Updater, error)
	NewProber(t Transport) (Prober, error)

	WithFileSystem(fs file.FileSystem) Operator
	WithDryRunMode(mode string) Operator
}

type Patch struct {
	Path  string
	Value any
}

type TransportContext interface {
	GetTransport() Transport
	GetFileSystem() file.FileSystem
	GetDryRunMode() string
}

type BuildOperator func() Operator

// RegisterOperator registers a factory function for operating Transport
// instances of a specific type.
// Note:
//   - The implementation package must be imported in internal/cmd/register.go;
//     otherwise, init() will not be called, the transport type will not be
//     registered, and an "unsupported connection type" error will occur at
//     runtime.
func RegisterOperator(transportType string, fn BuildOperator) {
	operatorFactories[transportType] = fn
}

var operatorFactories = map[string]BuildOperator{}

func NewOperator(transportType string) (Operator, error) {
	factory, ok := operatorFactories[transportType]
	if !ok {
		return nil, fmt.Errorf("unsupported transport type: %s", transportType)
	}
	return factory(), nil
}

func NewUpdater(transportType string, operation string, fs file.FileSystem) (Updater, error) {
	operator, err := NewOperator(transportType)
	if err != nil {
		return nil, err
	}
	return operator.WithFileSystem(fs).NewUpdater(operation)
}

func NewProber(ctx TransportContext) (Prober, error) {
	t := ctx.GetTransport()
	operator, err := NewOperator(t.GetType())
	if err != nil {
		return nil, err
	}
	return operator.
		WithFileSystem(ctx.GetFileSystem()).
		WithDryRunMode(ctx.GetDryRunMode()).
		NewProber(t)
}

func NewRemoteSession(ctx context.Context, tctx TransportContext) (RemoteSession, error) {
	t := tctx.GetTransport()
	operator, err := NewOperator(t.GetType())
	if err != nil {
		return nil, err
	}
	return operator.
		WithFileSystem(tctx.GetFileSystem()).
		WithDryRunMode(tctx.GetDryRunMode()).
		NewRemoteSession(ctx, t)
}
