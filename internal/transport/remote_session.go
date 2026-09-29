// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import (
	"context"
	"io"
	"os"
)

type RemoteSession interface {
	GetWorkRoot(ctx context.Context) (string, error)

	Upload(ctx context.Context, local, remote string, perm *os.FileMode) error
	Download(ctx context.Context, remote, local string, perm *os.FileMode) error
	DownloadDir(ctx context.Context, remote, local string, recursive bool) error
	MkdirAll(ctx context.Context, path string) error

	Execute(ctx context.Context, exec RemoteExecution) error

	Close() error
}

type RemoteExecution struct {
	Id         string
	LocalRoot  string
	RemoteRoot string
	Cmd        []string
	Workdir    string
	Envs       map[string]string
	Stdout     io.Writer
	Stderr     io.Writer
}
