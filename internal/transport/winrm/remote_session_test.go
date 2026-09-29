// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sony/niwashi/internal/transport"
)

// fakeClient is a test double for Client that records calls instead of
// talking WinRM. remoteSession's own logic (path composition, script
// construction) is what's under test here.
type fakeClient struct {
	uploaded     map[string][]byte
	downloadable map[string][]byte
	mkdirs       []string
	listDir      []DirEntry
	lastScript   string
	execErr      error
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		uploaded:     map[string][]byte{},
		downloadable: map[string][]byte{},
	}
}

func (c *fakeClient) Close() error { return nil }

func (c *fakeClient) RunWithContext(ctx context.Context, cmd string, stdout, stderr io.Writer) error {
	c.lastScript = cmd
	if strings.Contains(cmd, "$env:USERPROFILE") {
		_, _ = io.WriteString(stdout, `C:\Users\testuser`+"\r\n")
	}
	return c.execErr
}

func (c *fakeClient) MkdirAll(ctx context.Context, remotePath string) error {
	c.mkdirs = append(c.mkdirs, remotePath)
	return nil
}

func (c *fakeClient) UploadContent(ctx context.Context, remotePath string, data []byte) error {
	c.uploaded[remotePath] = append([]byte(nil), data...)
	return nil
}

func (c *fakeClient) DownloadContent(ctx context.Context, remotePath string) ([]byte, error) {
	data, ok := c.downloadable[remotePath]
	if !ok {
		return nil, fmt.Errorf("no such remote file: %s", remotePath)
	}
	return data, nil
}

func (c *fakeClient) ListDir(ctx context.Context, remotePath string) ([]DirEntry, error) {
	return c.listDir, nil
}

// memFS is a minimal in-memory file.FileSystem test double.
type memFS struct {
	files map[string][]byte
	dirs  []string
}

func newMemFS() *memFS {
	return &memFS{files: map[string][]byte{}}
}

func (f *memFS) Open(path string) (io.ReadSeekCloser, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &memReader{Reader: bytes.NewReader(data)}, nil
}

func (f *memFS) OpenFile(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	return &memWriter{fs: f, path: path}, nil
}

func (f *memFS) Stat(path string) (os.FileInfo, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return memFileInfo{size: int64(len(data))}, nil
}

func (f *memFS) MkdirAll(path string, perm os.FileMode) error {
	f.dirs = append(f.dirs, path)
	return nil
}

type memReader struct {
	*bytes.Reader
}

func (r *memReader) Close() error { return nil }

type memWriter struct {
	fs   *memFS
	path string
	buf  bytes.Buffer
}

func (w *memWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *memWriter) Close() error {
	w.fs.files[w.path] = w.buf.Bytes()
	return nil
}

type memFileInfo struct {
	size int64
}

func (i memFileInfo) Name() string       { return "" }
func (i memFileInfo) Size() int64        { return i.size }
func (i memFileInfo) Mode() os.FileMode  { return 0644 }
func (i memFileInfo) ModTime() time.Time { return time.Time{} }
func (i memFileInfo) IsDir() bool        { return false }
func (i memFileInfo) Sys() any           { return nil }

func newTestSession(t *testing.T, c *fakeClient, fs *memFS) *remoteSession {
	t.Helper()
	tp := &Transport{}
	tp.SetDefaults()
	return &remoteSession{
		fs:     fs,
		t:      tp,
		client: c,
	}
}

func TestRemoteSession_GetWorkRoot(t *testing.T) {
	s := newTestSession(t, newFakeClient(), newMemFS())

	got, err := s.GetWorkRoot(context.Background())
	if err != nil {
		t.Fatalf("GetWorkRoot() failed: %v", err)
	}
	want := `C:\Users\testuser`
	if got != want {
		t.Errorf("GetWorkRoot() = %q, want %q", got, want)
	}
}

func TestRemoteSession_Upload(t *testing.T) {
	fs := newMemFS()
	fs.files["local/script.ps1"] = []byte("Write-Host 'hi'")
	c := newFakeClient()
	s := newTestSession(t, c, fs)

	if err := s.Upload(context.Background(), "local/script.ps1", `C:\remote\script.ps1`, nil); err != nil {
		t.Fatalf("Upload() failed: %v", err)
	}

	got, ok := c.uploaded[`C:\remote\script.ps1`]
	if !ok {
		t.Fatalf("Upload() did not upload to expected remote path")
	}
	if string(got) != "Write-Host 'hi'" {
		t.Errorf("Upload() content = %q, want %q", got, "Write-Host 'hi'")
	}
}

func TestRemoteSession_Download(t *testing.T) {
	c := newFakeClient()
	c.downloadable[`C:\remote\out.log`] = []byte("output content")
	fs := newMemFS()
	s := newTestSession(t, c, fs)

	if err := s.Download(context.Background(), `C:\remote\out.log`, "local/out.log", nil); err != nil {
		t.Fatalf("Download() failed: %v", err)
	}

	got, ok := fs.files["local/out.log"]
	if !ok {
		t.Fatalf("Download() did not write to expected local path")
	}
	if string(got) != "output content" {
		t.Errorf("Download() content = %q, want %q", got, "output content")
	}
}

func TestRemoteSession_DownloadDir(t *testing.T) {
	c := newFakeClient()
	c.listDir = []DirEntry{
		{IsDir: true, RelativePath: "sub"},
		{IsDir: false, RelativePath: `sub\inner.txt`},
		{IsDir: false, RelativePath: "top.txt"},
	}
	c.downloadable[`C:\remote\sub\inner.txt`] = []byte("inner")
	c.downloadable[`C:\remote\top.txt`] = []byte("top")
	fs := newMemFS()
	s := newTestSession(t, c, fs)

	if err := s.DownloadDir(context.Background(), `C:\remote`, "local", true); err != nil {
		t.Fatalf("DownloadDir() failed: %v", err)
	}

	wantFiles := map[string]string{
		"local/sub/inner.txt": "inner",
		"local/top.txt":       "top",
	}
	for path, want := range wantFiles {
		got, ok := fs.files[path]
		if !ok {
			t.Errorf("DownloadDir() did not write %q", path)
			continue
		}
		if string(got) != want {
			t.Errorf("DownloadDir() content for %q = %q, want %q", path, got, want)
		}
	}

	foundSubDir := false
	for _, d := range fs.dirs {
		if d == "local/sub" {
			foundSubDir = true
		}
	}
	if !foundSubDir {
		t.Errorf("DownloadDir() did not create local dir for %q, dirs = %v", "sub", fs.dirs)
	}
}

func TestRemoteSession_Execute(t *testing.T) {
	c := newFakeClient()
	fs := newMemFS()
	s := newTestSession(t, c, fs)

	var stdout, stderr bytes.Buffer
	exec := transport.RemoteExecution{
		Id:         "task1",
		LocalRoot:  "local",
		RemoteRoot: `C:\Users\testuser\.niwashi`,
		Cmd:        []string{"powershell", `-File`, `C:\Users\testuser\.niwashi\logs\task1.ps1`},
		Workdir:    `C:\Users\testuser\.niwashi\recipe\test@1.0.0`,
		Envs:       map[string]string{"FOO": "it's a value"},
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	if err := s.Execute(context.Background(), exec); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	envPath := `C:\Users\testuser\.niwashi\logs\task1.ps1`
	envContent, ok := c.uploaded[envPath]
	if !ok {
		t.Fatalf("Execute() did not upload env file to %q", envPath)
	}
	if !strings.Contains(string(envContent), `$env:FOO = 'it''s a value'`) {
		t.Errorf("Execute() env file content = %q, missing expected assignment", envContent)
	}

	script := c.lastScript
	for _, want := range []string{
		". '" + envPath + "'",
		"Set-Location -LiteralPath '" + exec.Workdir + "'",
		"& 'powershell' -File '" + exec.Cmd[2] + "'",
		"exit $LASTEXITCODE",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("Execute() script = %q, missing expected fragment %q", script, want)
		}
	}
}

func TestRemoteSession_Execute_NonZeroExit(t *testing.T) {
	c := newFakeClient()
	c.execErr = fmt.Errorf("remote command exited with code 1")
	s := newTestSession(t, c, newMemFS())

	exec := transport.RemoteExecution{
		Id:         "task1",
		RemoteRoot: `C:\Users\testuser\.niwashi`,
		Cmd:        []string{"pwsh"},
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	if err := s.Execute(context.Background(), exec); err == nil {
		t.Fatal("Execute() succeeded unexpectedly for a failing remote command")
	}
}
