// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/workspace"
)

// recordingSession is a transport.RemoteSession test double that records
// the remote paths it's asked to create/upload to, so tests can assert on
// exactly what path strings the Uploader produced.
type recordingSession struct {
	mkdirs   []string
	uploaded []string
}

func (s *recordingSession) GetWorkRoot(ctx context.Context) (string, error) { return "", nil }

func (s *recordingSession) Upload(ctx context.Context, local, remote string, perm *os.FileMode) error {
	s.uploaded = append(s.uploaded, remote)
	return nil
}

func (s *recordingSession) Download(ctx context.Context, remote, local string, perm *os.FileMode) error {
	return nil
}

func (s *recordingSession) DownloadDir(ctx context.Context, remote, local string, recursive bool) error {
	return nil
}

func (s *recordingSession) MkdirAll(ctx context.Context, path string) error {
	s.mkdirs = append(s.mkdirs, path)
	return nil
}

func (s *recordingSession) Execute(ctx context.Context, exec transport.RemoteExecution) error {
	return nil
}

func (s *recordingSession) Close() error { return nil }

// backslashResolver joins with "\", standing in for a Windows remote target
// without depending on internal/transport/winrm.
type backslashResolver struct{}

func (backslashResolver) Find(key string) string { return "" }
func (backslashResolver) Join(paths ...string) string {
	out := paths[0]
	for _, p := range paths[1:] {
		out += `\` + p
	}
	return out
}
func (backslashResolver) Render(t string) string { return t }

func TestUploader_uploadDir_UsesResolverSeparator(t *testing.T) {
	localRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(localRoot, "sub"), 0750); err != nil {
		t.Fatalf("failed to set up local fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localRoot, "sub", "inner.txt"), []byte("inner"), 0600); err != nil {
		t.Fatalf("failed to set up local fixture file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localRoot, "top.txt"), []byte("top"), 0600); err != nil {
		t.Fatalf("failed to set up local fixture file: %v", err)
	}

	u := NewUploader(&mockFileSystem{}, backslashResolver{})
	u.AddDir(localRoot, `C:\remote\outputs`, true)

	session := &recordingSession{}
	if err := u.Upload(context.Background(), session); err != nil {
		t.Fatalf("Upload() failed: %v", err)
	}

	wantMkdir := `C:\remote\outputs\sub`
	if !contains(session.mkdirs, wantMkdir) {
		t.Errorf("MkdirAll calls = %v, want to contain %q", session.mkdirs, wantMkdir)
	}

	wantUploads := []string{`C:\remote\outputs\sub\inner.txt`, `C:\remote\outputs\top.txt`}
	for _, want := range wantUploads {
		if !contains(session.uploaded, want) {
			t.Errorf("Upload calls = %v, want to contain %q", session.uploaded, want)
		}
	}
}

func TestUploader_uploadRecipeAsset_UsesResolverSeparator(t *testing.T) {
	u := NewUploader(&mockFileSystem{}, backslashResolver{})
	u.AddRecipeAssets(&recipeSpec{
		fqid: "test/test@1.0.0",
		dir:  "testdata",
		assets: []string{
			"bin/setup.sh",
			"top.txt",
		},
	}, `C:\remote\recipe\test@1.0.0`)

	session := &recordingSession{}
	if err := u.Upload(context.Background(), session); err != nil {
		t.Fatalf("Upload() failed: %v", err)
	}

	wantMkdir := `C:\remote\recipe\test@1.0.0\bin`
	if !contains(session.mkdirs, wantMkdir) {
		t.Errorf("MkdirAll calls = %v, want to contain %q", session.mkdirs, wantMkdir)
	}

	wantUploads := []string{
		`C:\remote\recipe\test@1.0.0\bin\setup.sh`,
		`C:\remote\recipe\test@1.0.0\top.txt`,
	}
	for _, want := range wantUploads {
		if !contains(session.uploaded, want) {
			t.Errorf("Upload calls = %v, want to contain %q", session.uploaded, want)
		}
	}
}

func TestSplitRelPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "unix-style leading separator", in: "/sub/inner.txt", want: []string{"sub", "inner.txt"}},
		{name: "recipe asset style", in: "bin/setup.sh", want: []string{"bin", "setup.sh"}},
		{name: "single segment", in: "top.txt", want: []string{"top.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitRelPath(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitRelPath(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitRelPath(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

var _ workspace.PathResolver = backslashResolver{}
