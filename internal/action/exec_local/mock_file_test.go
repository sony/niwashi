// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"io"
	"os"
	"time"
)

type dummyReadWriteCloser struct{}

func (d *dummyReadWriteCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (d *dummyReadWriteCloser) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (d *dummyReadWriteCloser) Close() error {
	return nil
}

func (d *dummyReadWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// implement os.FileInfo interface
type dummyFileInfo struct{}

func (d *dummyFileInfo) Name() string       { return "dummy" }
func (d *dummyFileInfo) Size() int64        { return 0 }
func (d *dummyFileInfo) Mode() os.FileMode  { return 0644 }
func (d *dummyFileInfo) ModTime() time.Time { return time.Time{} }
func (d *dummyFileInfo) IsDir() bool        { return false }
func (d *dummyFileInfo) Sys() interface{}   { return nil }

// implement file.Creator interface
type mockFileCreator struct{}

func (c *mockFileCreator) OpenFile(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	return &dummyReadWriteCloser{}, nil
}

// implement file.Opener interface
type mockFileOpener struct{}

func (o *mockFileOpener) Open(path string) (io.ReadSeekCloser, error) {
	return &dummyReadWriteCloser{}, nil
}

// implement file.Statter interface
type mockFileStater struct{}

func (s *mockFileStater) Stat(path string) (os.FileInfo, error) {
	return &dummyFileInfo{}, nil
}

type mockDirCreator struct{}

func (c *mockDirCreator) MkdirAll(path string, _ os.FileMode) error {
	return nil
}

// implement file.FileSystem interface
type mockFileSystem struct {
	mockFileCreator
	mockFileOpener
	mockFileStater
	mockDirCreator
}
