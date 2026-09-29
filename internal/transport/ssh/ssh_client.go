// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SshClient interface {
	transport.Client
	NewSession() (SshSession, error)
	NewSftpClient() (SftpClient, error)
}

type sshClient struct {
	client *ssh.Client
}

func NewSshClient(ctx context.Context, c *Transport, fs file.FileSystem) (SshClient, error) {

	client, err := dialSsh(ctx, c, fs)
	if err != nil {
		return nil, err
	}

	return &sshClient{
		client: client,
	}, nil
}

func dialSsh(ctx context.Context, c *Transport, fs file.FileSystem) (*ssh.Client, error) {

	signer, err := c.Auth.CreateSigner(fs)
	if err != nil {
		return nil, err
	}

	hostKeyCallback, err := knownhosts.New(c.HostKey.KnownHostsPath)
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User:            c.Address.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
	}

	addr := net.JoinHostPort(c.Address.Host, fmt.Sprintf("%d", c.Address.Port))

	dialTimeout := c.Options.ConnectTimeout()
	dialer := &net.Dialer{Timeout: dialTimeout}

	var conn net.Conn

	maxRetries := c.Options.RetryMax()
	retryInterval := c.Options.RetryInterval()

	for attempt := 0; ; attempt++ {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			break
		}

		if attempt > maxRetries-1 {
			return nil, errors.Join(err, fmt.Errorf("failed to connect to %s after %d attempts", addr, maxRetries))
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
			// Try again after waiting for the retry interval
		}
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		hsTimeout := c.Options.HandshakeTimeout()
		_ = conn.SetDeadline(time.Now().Add(hsTimeout))
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	_ = conn.SetDeadline(time.Time{})

	return ssh.NewClient(clientConn, chans, reqs), nil
}

func (c *sshClient) NewSession() (SshSession, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	return &sshSession{
		session: session,
	}, nil
}

func (c *sshClient) NewSftpClient() (SftpClient, error) {
	return NewSftpClient(c.client)
}

func (c *sshClient) Close() error {
	return c.client.Close()
}
