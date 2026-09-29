// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"fmt"
	"os"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/workspace"
)

const knownHostsRelative = "ssh/known_hosts"

const rootDirVar = "{{ .Paths.workspace }}"

type KnownHostsUpdater struct {
	operation  string
	fs         file.FileSystem
	ws         *workspace.RunWorkspace
	systemPath string
}

func NewKnownHostsUpdater(operation string, fs file.FileSystem) *KnownHostsUpdater {
	return &KnownHostsUpdater{
		operation: operation,
		fs:        fs,
	}
}

func (u *KnownHostsUpdater) Update(t transport.Transport, ws *workspace.RunWorkspace) ([]transport.Patch, error) {
	// do nothing if no hostKey field
	tp, ok := t.(*Transport)
	if !ok {
		return nil, fmt.Errorf("invalid transport type for ssh.KnownHostsUpdater")
	}

	u.ws = ws
	u.systemPath = ws.Join(rootDirVar, knownHostsRelative)

	switch u.operation {
	case transport.OperationConstruct:
		return u.copyHostEntry(tp)
	case transport.OperationDestruct:
		return u.removeHostEntry(tp)
	default:
		return nil, fmt.Errorf("unsupported operation in ssh.UpdateTransport")
	}
}

func getKnwonHostsPath(ws *workspace.RootWorkspace) string {
	return ws.Join(ws.GetRootDirPath(), knownHostsRelative)
}

func (u *KnownHostsUpdater) copyHostEntry(tp *Transport) ([]transport.Patch, error) {

	// do nothing if no hostKey field or system path set already
	if tp.HostKey == nil || tp.HostKey.KnownHostsPath == u.systemPath {
		// nothing to do
		return nil, nil
	}

	// Copy host info from original known_hosts to system path
	nh := NewKnownHosts(u.fs)

	knownHostsPath := getKnwonHostsPath(u.ws.RootWorkspace)

	// raise error if failed to load known_hosts file, but ignore if file not exist (will create new file later)
	if err := nh.LoadHostEntries(knownHostsPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load known hosts: %w", err)
	}
	if err := nh.CopyHostEntry(tp); err != nil {
		return nil, fmt.Errorf("failed to copy host entry: %w", err)
	}

	if err := nh.SaveToFile(knownHostsPath, true); err != nil {
		return nil, fmt.Errorf("failed to save known hosts: %w", err)
	}

	// change knownHostsPath to system known_hosts path
	return []transport.Patch{
		{
			Path:  "/hostKey/knownHostsPath",
			Value: u.systemPath,
		},
	}, nil
}

func (u *KnownHostsUpdater) removeHostEntry(tp *Transport) ([]transport.Patch, error) {

	// do nothing if no hostKey field or system path not set
	if tp.HostKey == nil || tp.HostKey.KnownHostsPath != u.systemPath {
		// nothing to do
		return nil, nil
	}

	knownHostsPath := getKnwonHostsPath(u.ws.RootWorkspace)
	nh, err := NewKnownHostsFromFile(u.fs, knownHostsPath)
	// raise error if failed to load known_hosts file
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("known_hosts file not found, skipping host entry removal:", "knownHostsPath", knownHostsPath)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load known hosts: %w", err)
	}
	if err := nh.RemoveHostEntry(tp); err != nil {
		return nil, fmt.Errorf("failed to remove host entry: %w", err)
	}
	if err := nh.SaveToFile(knownHostsPath, false); err != nil {
		return nil, fmt.Errorf("failed to save known hosts: %w", err)
	}

	// don't change path
	return nil, nil
}
