// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"io"
	"os"
	"path/filepath"
)

type Opener interface {
	Open(path string) (io.ReadSeekCloser, error)
}

type LocalOpener struct{}

func (o *LocalOpener) Open(path string) (io.ReadSeekCloser, error) {
	// #nosec G304 -- path is operator-controlled (CLI args or recipe definitions).
	// filepath.Clean is applied to mitigate path traversal. No untrusted input is processed.
	return os.Open(filepath.Clean(path))
}
