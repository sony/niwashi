// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"errors"
	"io"
	"os"
)

type FileSystem interface {
	Opener
	Creator
	Statter
	DirCreator
}

type LocalFileSystem struct {
	LocalCreator
	LocalOpener
	LocalStatter
	LocalDirCreator
}

func NewDefaultFileSystem() FileSystem {
	return &LocalFileSystem{}
}

func ReadFile(fs Opener, filepath string) ([]byte, error) {
	file, err := fs.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	return io.ReadAll(file)
}

func WriteFile(f Creator, filepath string, data []byte, perm os.FileMode) (err error) {
	file, err := f.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	_, err = file.Write(data)
	return err
}
