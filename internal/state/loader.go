// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"
	"io"

	"github.com/sony/niwashi/internal/file"
)

func load(r io.ReadSeeker, f func(any, io.Reader) error) (*State, error) {
	hash, err := file.GenerateHash(r)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var state State
	err = f(&state, r)
	if err != nil {
		return nil, err
	}

	state.Hash = hash

	err = state.Validate()
	if err != nil {
		return nil, err
	}

	state.SetDefaults()

	return &state, nil
}

func LoadAsJson(r io.ReadSeeker) (*State, error) {
	return load(r, file.LoadAsJson)
}

func LoadAsYaml(r io.ReadSeeker) (*State, error) {
	return load(r, file.LoadAsYaml)
}

func Load(p string, opener file.Opener) (*State, error) {
	var state State
	var hash string
	err := file.LoadWithDecoding(&state, p, &hash, opener)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to load state from: %q", p))
	}

	state.Hash = hash

	// todo: validate
	state.SetDefaults()

	return &state, nil
}
