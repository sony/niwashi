// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"io"
	"os"
	"path/filepath"
)

type Creator interface {
	OpenFile(path string, flag int, perm os.FileMode) (io.WriteCloser, error)
}

type LocalCreator struct{}

func (c *LocalCreator) OpenFile(path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- path is operator-controlled (CLI args or recipe definitions).
	// filepath.Clean is applied to mitigate path traversal. No untrusted input is processed.
	return os.OpenFile(filepath.Clean(path), flag, perm)
}
