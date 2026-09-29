// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/transport/ssh"
)

type dummyOperator struct {
	History []string
}

func (d *dummyOperator) NewClient(ctx context.Context, t transport.Transport) (transport.Client, error) {
	tp := t.(*ssh.Transport)
	if tp == nil {
		return nil, fmt.Errorf("invalid transport type for dummyFactory")
	}

	return &mockSshClient{
		parent: d,
	}, nil
}

func (d *dummyOperator) NewUpdater(op string) (transport.Updater, error) {
	return nil, nil
}

func (d *dummyOperator) NewProber(t transport.Transport) (transport.Prober, error) {
	return nil, nil
}

func (d *dummyOperator) NewRemoteSession(ctx context.Context, t transport.Transport) (transport.RemoteSession, error) {
	return &mockRemoteSession{}, nil
}

func (d *dummyOperator) WithFileSystem(fs file.FileSystem) transport.Operator {
	return d
}

func (d *dummyOperator) WithDryRunMode(mode string) transport.Operator {
	return d
}

type mockSshClient struct {
	parent *dummyOperator
}

func (m *mockSshClient) NewSession() (ssh.SshSession, error) {
	return &mockSshSession{
		parent: m.parent,
	}, nil
}

func (m *mockSshClient) NewSftpClient() (ssh.SftpClient, error) {
	return &mockSftpClient{
		parent: m.parent,
	}, nil
}

func (m *mockSshClient) Close() error {
	return nil
}

type mockSshSession struct {
	parent *dummyOperator
}

func (m *mockSshSession) Output(cmd string) ([]byte, error) {
	m.parent.History = append(m.parent.History, "Output: "+cmd)
	if cmd == "echo $HOME" {
		return []byte("/home/testuser\n"), nil
	}
	return []byte("mock output"), nil
}

func (m *mockSshSession) Run(cmd string) error {
	m.parent.History = append(m.parent.History, "Run: "+cmd)
	return nil
}

func (m *mockSshSession) Shell() error {
	m.parent.History = append(m.parent.History, "Shell")
	return nil
}
func (m *mockSshSession) Wait() error {
	m.parent.History = append(m.parent.History, "Wait")
	return nil
}

func (m *mockSshSession) Close() error {
	return nil
}

func (m *mockSshSession) SetStdout(stdout io.Writer) {
}

func (m *mockSshSession) SetStderr(stderr io.Writer) {
}
func (m *mockSshSession) SetStdin(stdin io.Reader) {
}

type mockSftpClient struct {
	parent *dummyOperator
}

func (m *mockSftpClient) MkdirAll(path string) error {
	m.parent.History = append(m.parent.History, "MkdirAll: "+path)
	return nil
}

func (m *mockSftpClient) Create(path string) (io.WriteCloser, error) {
	m.parent.History = append(m.parent.History, "Create: "+path)
	return &mockReadWriteCloser{}, nil
}

func (m *mockSftpClient) Open(path string) (io.ReadCloser, error) {
	m.parent.History = append(m.parent.History, "Open: "+path)
	return &mockReadWriteCloser{}, nil
}

func (m *mockSftpClient) Chmod(path string, perm os.FileMode) error {
	m.parent.History = append(m.parent.History, "Chmod: "+path)
	return nil
}

func (m *mockSftpClient) Walk(root string) ssh.Walker {
	m.parent.History = append(m.parent.History, "Walk: "+root)
	return &mockWalker{}
}

func (m *mockSftpClient) Stat(path string) (os.FileInfo, error) {
	m.parent.History = append(m.parent.History, "Stat: "+path)
	return nil, nil
}

func (m *mockSftpClient) Close() error {
	return nil
}

type mockWalker struct{}

func (w *mockWalker) Err() error {
	return nil
}

func (w *mockWalker) Path() string {
	return "a"
}

func (w *mockWalker) SkipDir() {
}

func (w *mockWalker) Stat() os.FileInfo {
	return nil
}

func (w *mockWalker) Step() bool {
	return false
}

type mockReadWriteCloser struct{}

func (m *mockReadWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockReadWriteCloser) Read(p []byte) (n int, err error) {
	copy(p, []byte("mock file content"))
	return len("mock file content"), io.EOF
}

func (m *mockReadWriteCloser) Close() error {
	return nil
}

type mockRemoteSession struct{}

func (m *mockRemoteSession) GetWorkRoot(ctx context.Context) (string, error) {
	return "/tmp/xxx", nil
}

func (m *mockRemoteSession) Upload(ctx context.Context, local, remote string, perm *os.FileMode) error {
	return nil
}

func (m *mockRemoteSession) Download(ctx context.Context, remote, local string, perm *os.FileMode) error {
	return nil
}

func (m *mockRemoteSession) DownloadDir(ctx context.Context, remote, local string, recursive bool) error {
	return nil
}

func (m *mockRemoteSession) MkdirAll(ctx context.Context, path string) error {
	return nil
}

func (m *mockRemoteSession) Execute(ctx context.Context, exec transport.RemoteExecution) error {
	return nil
}

func (m *mockRemoteSession) Close() error {
	return nil
}
