// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/transport"
)

type remoteSession struct {
	fs         file.FileSystem
	t          *Transport
	client     SshClient
	sftpClient SftpClient
}

func NewRemoteSession(c transport.Transport, fs file.FileSystem) *remoteSession {
	t := c.(*Transport)

	return &remoteSession{
		fs: fs,
		t:  t,
	}
}

func (s *remoteSession) getClient(ctx context.Context) (SshClient, error) {
	if s.client != nil {
		return s.client, nil
	}

	client, err := NewSshClient(ctx, s.t, s.fs)
	if err != nil {
		return nil, err
	}

	s.client = client
	return s.client, nil
}

func (s *remoteSession) getSftpClient(ctx context.Context) (SftpClient, error) {
	if s.sftpClient != nil {
		return s.sftpClient, nil
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	sftpClient, err := client.NewSftpClient()
	if err != nil {
		return nil, err
	}

	s.sftpClient = sftpClient
	return s.sftpClient, nil
}

func (s *remoteSession) GetWorkRoot(ctx context.Context) (_ string, err error) {

	client, err := s.getClient(ctx)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = session.Close()
	}()

	o, err := session.Output("echo $HOME")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(o)), nil
}

func (s *remoteSession) Upload(ctx context.Context, local, remote string, perm *os.FileMode) (err error) {

	client, err := s.getSftpClient(ctx)
	if err != nil {
		return err
	}

	logger.Info("Uploading file", "local", local, "remote", remote)

	src, err := s.fs.Open(local)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to open local file for upload: %s", local))
	}
	defer func() {
		if e := src.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to close local file: %s", local), e)
		}
	}()

	dest, err := client.Create(remote)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to create remote file: %s", remote))
	}
	defer func() {
		if e := dest.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to close remote file: %s", remote), e)
		}
	}()

	if _, err = io.Copy(dest, src); err != nil {
		return errors.Join(err, fmt.Errorf("failed to copy file content to remote: %s", remote))
	}

	var mode os.FileMode
	if perm == nil {
		fi, err := s.fs.Stat(local)
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to stat local file: %s", local))
		}
		mode = fi.Mode()
	} else {
		mode = *perm
	}

	// set permission
	if err := client.Chmod(remote, mode); err != nil {
		return errors.Join(err, fmt.Errorf("failed to set permission for remote file: %s", remote))
	}

	return nil
}

func (s *remoteSession) Download(ctx context.Context, remote, local string, perm *os.FileMode) error {

	client, err := s.getSftpClient(ctx)
	if err != nil {
		return err
	}

	return s.downloadFile(ctx, client, remote, local, perm)
}

func (s *remoteSession) downloadFile(
	ctx context.Context, client SftpClient, remote, local string, perm *os.FileMode) (err error) {

	logger.Info("Downloading file from remote:", "remote", remote, "local", local)

	srcFile, err := client.Open(remote)
	if err != nil {
		return err
	}
	defer func() {
		if e := srcFile.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to close remote file: %s", remote), e)
		}
	}()

	p := perm
	if perm == nil {
		fi, err := client.Stat(remote)
		if err != nil {
			return err
		}
		mode := fi.Mode()
		p = &mode
	}

	dstFile, err := s.fs.OpenFile(local, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, *p)
	if err != nil {
		return err
	}
	defer func() {
		if e := dstFile.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to close local file: %s", local), e)
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *remoteSession) DownloadDir(
	ctx context.Context, remote, local string, recursive bool) error {

	client, err := s.getSftpClient(ctx)
	if err != nil {
		return err
	}

	walker := client.Walk(remote)
	for walker.Step() {
		if walker.Err() != nil {
			return walker.Err()
		}

		remotePath := walker.Path()
		relPath, err := filepath.Rel(remote, remotePath)
		if err != nil {
			return err
		}
		localPath := filepath.Join(local, relPath)

		fi := walker.Stat()
		if fi.IsDir() {
			// create local dir
			if err := s.fs.MkdirAll(localPath, fi.Mode()); err != nil {
				return err
			}
			continue
		}

		mode := fi.Mode()
		if err := s.downloadFile(ctx, client, remotePath, localPath, &mode); err != nil {
			return err
		}
	}

	return nil
}

func (s *remoteSession) MkdirAll(ctx context.Context, remote string) error {
	client, err := s.getSftpClient(ctx)
	if err != nil {
		return err
	}
	logger.Info("remote: mkdir", "remote", remote)
	return client.MkdirAll(remote)
}

func (s *remoteSession) Execute(ctx context.Context, exec transport.RemoteExecution) (err error) {
	remoteEnv, err := s.generateEnvFile(ctx, &exec)
	if err != nil {
		return err
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf(". %s && cd %s && %s",
		remoteEnv,
		exec.Workdir,
		strings.Join(exec.Cmd, " "))

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	session.SetStdout(exec.Stdout)
	session.SetStderr(exec.Stderr)

	if err := session.Run(cmd); err != nil {
		return err
	}

	return nil
}

func (s *remoteSession) Close() error {
	var err error
	if s.sftpClient != nil {
		err = errors.Join(err, s.sftpClient.Close())
		s.sftpClient = nil
	}
	if s.client != nil {
		err = errors.Join(err, s.client.Close())
		s.client = nil
	}

	return err
}

func (s *remoteSession) generateEnvFile(ctx context.Context, exec *transport.RemoteExecution) (string, error) {

	local := filepath.Join(exec.LocalRoot, "logs", exec.Id+".env")
	remote := filepath.Join(exec.RemoteRoot, "logs", exec.Id+".env")

	// generate local envfile content

	shellescape := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	}

	var sb strings.Builder
	for k, v := range exec.Envs {
		fmt.Fprintf(&sb, "export %s=%s\n", k, shellescape(v))
	}

	if err := s.WriteFile(local, []byte(sb.String()), 0644); err != nil {
		return "", err
	}

	if err := s.Upload(ctx, local, remote, nil); err != nil {
		return "", err
	}

	return remote, nil
}

func (s *remoteSession) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	w, err := s.fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		if e := w.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to close file: %s", path), e)
		}
	}()

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
