// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"
	"io"

	"github.com/sony/niwashi/internal/file"
)

func (s *State) WriteAsJson(w io.Writer) error {
	err := file.WriteAsJson(s, w)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to write state as JSON"))
	}

	return nil
}

func (s *State) WriteAsYaml(w io.Writer) error {
	err := file.WriteAsYaml(s, w)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to write state as YAML"))
	}
	return nil
}

func (s *State) WriteToFile(p string, fs file.Creator) error {
	err := file.WriteWithEncoding(s, p, fs)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to write state to: %q", p))
	}

	return nil
}
