// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SftpClient interface {
	MkdirAll(path string) error
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadCloser, error)
	Chmod(path string, perm os.FileMode) error
	Walk(root string) Walker
	Stat(path string) (os.FileInfo, error)

	Close() error
}

type Walker interface {
	Err() error
	Path() string
	SkipDir()
	Stat() os.FileInfo
	Step() bool
}

type sftpClient struct {
	client *sftp.Client
}

func NewSftpClient(client *ssh.Client) (SftpClient, error) {
	c, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}

	return &sftpClient{
		client: c,
	}, nil
}

func (c *sftpClient) MkdirAll(path string) error {
	return c.client.MkdirAll(path)
}

func (c *sftpClient) Create(path string) (io.WriteCloser, error) {
	return c.client.Create(path)
}

func (c *sftpClient) Chmod(path string, perm os.FileMode) error {
	return c.client.Chmod(path, perm)
}

func (c *sftpClient) Close() error {
	return c.client.Close()
}

func (c *sftpClient) Walk(root string) Walker {
	return c.client.Walk(root)
}

func (c *sftpClient) Open(path string) (io.ReadCloser, error) {
	return c.client.Open(path)
}

func (c *sftpClient) Stat(path string) (os.FileInfo, error) {
	return c.client.Stat(path)
}
