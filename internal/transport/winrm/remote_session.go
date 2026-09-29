// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/transport"
)

// remoteRootSeparator is hardcoded to "\" (rather than taken from
// exec.RemoteRoot's caller-chosen separator) because this package only ever
// talks to Windows targets: WinRM is a Windows-only protocol.
const remoteRootSeparator = "\\"

type remoteSession struct {
	fs     file.FileSystem
	t      *Transport
	client Client
}

func NewRemoteSession(c transport.Transport, fs file.FileSystem) *remoteSession {
	t := c.(*Transport)

	return &remoteSession{
		fs: fs,
		t:  t,
	}
}

func (s *remoteSession) getClient() (Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	c, err := NewClient(s.t, s.fs)
	if err != nil {
		return nil, err
	}

	s.client = c
	return s.client, nil
}

func (s *remoteSession) GetWorkRoot(ctx context.Context) (string, error) {
	client, err := s.getClient()
	if err != nil {
		return "", err
	}

	var stdout, stderr strings.Builder
	if err := client.RunWithContext(ctx, "$env:USERPROFILE", &stdout, &stderr); err != nil {
		return "", fmt.Errorf("failed to get remote user profile directory: %w (%s)", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (s *remoteSession) Upload(ctx context.Context, local, remote string, perm *os.FileMode) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	logger.Info("Uploading file", "local", local, "remote", remote)

	data, err := file.ReadFile(s.fs, local)
	if err != nil {
		return fmt.Errorf("failed to open local file for upload: %s: %w", local, err)
	}

	// perm is not applied: Windows/NTFS has no equivalent of the POSIX
	// executable bit, and PowerShell scripts are always invoked explicitly
	// (e.g. "pwsh -File script.ps1"), never via a bare "./script".
	return client.UploadContent(ctx, remote, data)
}

func (s *remoteSession) Download(ctx context.Context, remote, local string, perm *os.FileMode) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	return s.downloadFile(ctx, client, remote, local, perm)
}

func (s *remoteSession) downloadFile(ctx context.Context, client Client, remote, local string, perm *os.FileMode) error {
	logger.Info("Downloading file from remote:", "remote", remote, "local", local)

	data, err := client.DownloadContent(ctx, remote)
	if err != nil {
		return err
	}

	mode := os.FileMode(0644)
	if perm != nil {
		mode = *perm
	}

	return file.WriteFile(s.fs, local, data, mode)
}

func (s *remoteSession) DownloadDir(ctx context.Context, remote, local string, recursive bool) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	entries, err := client.ListDir(ctx, remote)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		segments := strings.Split(entry.RelativePath, remoteRootSeparator)
		localPath := filepath.Join(append([]string{local}, segments...)...)

		if entry.IsDir {
			if !recursive {
				continue
			}
			if err := s.fs.MkdirAll(localPath, 0755); err != nil {
				return err
			}
			continue
		}

		if !recursive && strings.Contains(entry.RelativePath, remoteRootSeparator) {
			// nested file, but caller asked for a flat (non-recursive) copy
			continue
		}

		remotePath := remote + remoteRootSeparator + entry.RelativePath
		if err := s.downloadFile(ctx, client, remotePath, localPath, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *remoteSession) MkdirAll(ctx context.Context, remote string) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	logger.Info("remote: mkdir", "remote", remote)
	return client.MkdirAll(ctx, remote)
}

func (s *remoteSession) Execute(ctx context.Context, exec transport.RemoteExecution) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	envPath, err := s.generateEnvFile(ctx, client, &exec)
	if err != nil {
		return err
	}

	// quote command arguments for PowerShell except '-File'
	cmdParts := make([]string, len(exec.Cmd))
	for i, arg := range exec.Cmd {
		if arg != "-File" {
			cmdParts[i] = psQuote(arg)
		} else {
			cmdParts[i] = arg
		}
	}

	lines := []string{
		"$ErrorActionPreference = 'Stop'",
		fmt.Sprintf(". %s", psQuote(envPath)),
		fmt.Sprintf("Set-Location -LiteralPath %s", psQuote(exec.Workdir)),
		fmt.Sprintf("& %s", strings.Join(cmdParts, " ")),
		"exit $LASTEXITCODE",
	}

	return client.RunWithContext(ctx, strings.Join(lines, "\n"), exec.Stdout, exec.Stderr)
}

func (s *remoteSession) Close() error {
	if s.client != nil {
		err := s.client.Close()
		s.client = nil
		return err
	}
	return nil
}

func (s *remoteSession) generateEnvFile(ctx context.Context, client Client, exec *transport.RemoteExecution) (string, error) {
	var sb strings.Builder
	for k, v := range exec.Envs {
		fmt.Fprintf(&sb, "$env:%s = %s\n", k, psQuote(v))
	}

	remote := exec.RemoteRoot + remoteRootSeparator + "logs" + remoteRootSeparator + exec.Id + ".ps1"
	if err := client.UploadContent(ctx, remote, []byte(sb.String())); err != nil {
		return "", errors.Join(err, fmt.Errorf("failed to upload env file: %s", remote))
	}

	return remote, nil
}
