// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import "os"

type DirCreator interface {
	MkdirAll(path string, perm os.FileMode) error
}

type LocalDirCreator struct{}

func (c *LocalDirCreator) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func MakeDirsAll(fs DirCreator, dirs ...string) error {

	for _, dir := range dirs {
		if err := makeDirIfNotExists(dir, fs); err != nil {
			return err
		}
	}

	return nil
}

func makeDirIfNotExists(dir string, fs DirCreator) error {
	if err := fs.MkdirAll(dir, 0o755); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
