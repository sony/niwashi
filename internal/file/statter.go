// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import "os"

type Statter interface {
	Stat(path string) (os.FileInfo, error)
}

type LocalStatter struct{}

func (s *LocalStatter) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
