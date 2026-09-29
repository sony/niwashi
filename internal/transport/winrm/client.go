// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	mwinrm "github.com/masterzen/winrm"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/transport"
)

// maxUploadSize is a conservative limit on how large a single file this
// package will upload in one command. WinRM has no SFTP-like streaming
// transfer: content is base64-encoded into a single PowerShell command, and
// both the client's SOAP envelope size and the server's MaxEnvelopeSizekb
// setting cap how large that command can be. Chunked upload (as
// Ansible/winrmcp do for arbitrarily large files) isn't implemented in v1
// (design proposal 026) -- oversized files are rejected with a clear error
// instead of failing deep inside the WinRM protocol layer.
//
// Backlog for a future revision: WinRM upload is inherently slow (one round
// trip per chunk) even with chunking. Bypassing WinRM for large transfers --
// e.g. have the target node pull the file itself (Invoke-WebRequest against
// a short-lived HTTP server on the host) rather than pushing it over WinRM
// commands -- is reportedly both simpler and much faster than chunked
// upload, and is worth evaluating before implementing chunking here.
const maxUploadSize = 100_000 // bytes, before base64 encoding

// DirEntry is one entry returned by Client.ListDir.
type DirEntry struct {
	// RelativePath is relative to the queried root, using "\" as separator
	// (as reported by PowerShell's FullName).
	RelativePath string
	IsDir        bool
}

// Client is the low-level WinRM client used by remoteSession and Prober. It
// wraps github.com/masterzen/winrm and adds the file-transfer helpers WinRM
// itself doesn't provide (there is no WinRM equivalent of SFTP).
type Client interface {
	transport.Client

	RunWithContext(ctx context.Context, cmd string, stdout, stderr io.Writer) error
	MkdirAll(ctx context.Context, remotePath string) error
	UploadContent(ctx context.Context, remotePath string, data []byte) error
	DownloadContent(ctx context.Context, remotePath string) ([]byte, error)
	ListDir(ctx context.Context, remotePath string) ([]DirEntry, error)
}

// winrmRunner is the subset of *masterzen/winrm.Client's API this package
// depends on. Depending on this narrow interface instead of *mwinrm.Client
// directly lets tests substitute a fake that doesn't speak the WinRM wire
// protocol: this package's own logic (command construction, exit-code
// handling, chunking limits, path parsing) is what's under test here, not
// masterzen/winrm's protocol implementation, which is out of scope per
// design proposal 026's testing strategy.
type winrmRunner interface {
	RunWithContext(ctx context.Context, cmd string, stdout, stderr io.Writer) (int, error)
}

type client struct {
	winrm winrmRunner
}

// NewClient creates a WinRM client for t. This doesn't connect; masterzen/winrm
// establishes a connection lazily, per command. fs is used to resolve
// t.Auth.PasswordRef.FromFile, if set.
func NewClient(t *Transport, fs file.Opener) (Client, error) {
	if t.Options.AllowInsecureHTTP {
		logger.Warn(
			"winrm transport is using plain HTTP instead of HTTPS; credentials and command output are not protected in transit",
			"host", t.Address.Host,
		)
	}

	password, err := t.Auth.GetPassword(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve winrm password: %w", err)
	}

	var caCert []byte
	if t.ServerCert.CACertPath != "" {
		caCert, err = file.ReadFile(fs, t.ServerCert.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read winrm CA cert %q: %w", t.ServerCert.CACertPath, err)
		}
	}

	endpoint := mwinrm.NewEndpoint(
		t.Address.Host,
		t.Address.Port,
		t.Options.UseHTTPS(),
		t.ServerCert.InsecureSkipVerify,
		caCert, nil, nil,
		time.Duration(t.Options.ConnectTimeoutSec)*time.Second,
	)

	params := mwinrm.DefaultParameters
	params.TransportDecorator = func() mwinrm.Transporter {
		return &mwinrm.ClientNTLM{}
	}

	c, err := mwinrm.NewClientWithParameters(endpoint, t.Address.User, password, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create winrm client: %w", err)
	}

	return &client{winrm: c}, nil
}

func (c *client) Close() error {
	// masterzen/winrm has no persistent connection to close; every command
	// creates and deletes its own WinRM shell.
	return nil
}

// RunWithContext runs cmd as a PowerShell script. WinRM's default shell for
// a plain command line is cmd.exe, not PowerShell, so cmd is wrapped via
// mwinrm.Powershell (base64-encoded "powershell.exe -EncodedCommand ...")
// before being sent -- every command this package builds (env setup,
// New-Item, [System.IO.File]::..., etc.) is PowerShell syntax.
func (c *client) RunWithContext(ctx context.Context, cmd string, stdout, stderr io.Writer) error {
	encoded := mwinrm.Powershell(cmd)
	if encoded == "" {
		return fmt.Errorf("failed to encode PowerShell command")
	}

	exitCode, err := c.winrm.RunWithContext(ctx, encoded, stdout, stderr)
	if err != nil {
		return fmt.Errorf("failed to run remote command: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("remote command exited with code %d", exitCode)
	}
	return nil
}

func (c *client) MkdirAll(ctx context.Context, remotePath string) error {
	cmd := fmt.Sprintf(`New-Item -ItemType Directory -Force -Path '%s' | Out-Null`, psEscape(remotePath))
	var stderr strings.Builder
	if err := c.RunWithContext(ctx, cmd, io.Discard, &stderr); err != nil {
		return fmt.Errorf("failed to create remote directory %q: %w (%s)", remotePath, err, stderr.String())
	}
	return nil
}

func (c *client) UploadContent(ctx context.Context, remotePath string, data []byte) error {
	if len(data) > maxUploadSize {
		return fmt.Errorf(
			"file too large to upload over winrm in a single command (%d bytes, limit %d bytes): %s",
			len(data), maxUploadSize, remotePath)
	}

	cmd := fmt.Sprintf(
		`[System.IO.File]::WriteAllBytes('%s', [System.Convert]::FromBase64String('%s'))`,
		psEscape(remotePath), base64.StdEncoding.EncodeToString(data),
	)
	var stderr strings.Builder
	if err := c.RunWithContext(ctx, cmd, io.Discard, &stderr); err != nil {
		return fmt.Errorf("failed to upload file to %q: %w (%s)", remotePath, err, stderr.String())
	}
	return nil
}

func (c *client) DownloadContent(ctx context.Context, remotePath string) ([]byte, error) {
	cmd := fmt.Sprintf(
		`[System.Convert]::ToBase64String([System.IO.File]::ReadAllBytes('%s'))`,
		psEscape(remotePath),
	)
	var stdout, stderr strings.Builder
	if err := c.RunWithContext(ctx, cmd, &stdout, &stderr); err != nil {
		return nil, fmt.Errorf("failed to download file from %q: %w (%s)", remotePath, err, stderr.String())
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(stdout.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to decode downloaded content from %q: %w", remotePath, err)
	}
	return decoded, nil
}

const dirEntryFieldSep = "|"

func (c *client) ListDir(ctx context.Context, remotePath string) ([]DirEntry, error) {
	// One line per entry: "<0|1 (is dir)>|<path relative to remotePath, \-separated>"
	cmd := fmt.Sprintf(
		`$base = (Resolve-Path -LiteralPath '%s').Path; `+
			`Get-ChildItem -LiteralPath '%s' -Recurse | ForEach-Object { "{0}|{1}" -f [int]$_.PSIsContainer, $_.FullName.Substring($base.Length + 1) }`,
		psEscape(remotePath), psEscape(remotePath),
	)
	var stdout, stderr strings.Builder
	if err := c.RunWithContext(ctx, cmd, &stdout, &stderr); err != nil {
		return nil, fmt.Errorf("failed to list remote directory %q: %w (%s)", remotePath, err, stderr.String())
	}

	var entries []DirEntry
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		isDir, relPath, ok := strings.Cut(line, dirEntryFieldSep)
		if !ok {
			continue
		}
		entries = append(entries, DirEntry{
			IsDir:        isDir == "1",
			RelativePath: relPath,
		})
	}
	return entries, nil
}

// psEscape escapes a value for embedding in a PowerShell single-quoted
// string literal.
func psEscape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// psQuote wraps a value as a single-quoted PowerShell string literal.
func psQuote(s string) string {
	return "'" + psEscape(s) + "'"
}
