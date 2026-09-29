// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"io"
	"os"
)

type File interface {
	Name() string
	Open() (io.ReadSeekCloser, error)
	OpenFile(flag int, perm os.FileMode) (io.WriteCloser, error)
	Stat() (os.FileInfo, error)
}

func NewFile(path string, fs FileSystem) File {
	if fs == nil {
		fs = NewDefaultFileSystem()
	}
	return &fsFile{
		name: path,
		fs:   fs,
	}
}

type fsFile struct {
	name string
	fs   FileSystem
}

func (f *fsFile) Name() string {
	return f.name
}

func (f *fsFile) Open() (io.ReadSeekCloser, error) {
	return f.fs.Open(f.name)
}

func (f *fsFile) OpenFile(flag int, perm os.FileMode) (io.WriteCloser, error) {
	return f.fs.OpenFile(f.name, flag, perm)
}

func (f *fsFile) Stat() (os.FileInfo, error) {
	return f.fs.Stat(f.name)
}
