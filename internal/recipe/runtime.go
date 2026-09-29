// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"
	"slices"

	"github.com/sony/niwashi/internal/types"
)

var (
	RuntimeTypes = []string{
		"tool",
		"service",
	}
)

type Runtime struct {
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

func (r *Runtime) Validate() error {
	if r == nil {
		return nil
	}

	var err error

	if !slices.Contains(RuntimeTypes, r.Type) {
		err = errors.Join(err, fmt.Errorf("invalid runtime type: %s", r.Type))
	}

	if r.Name == "" {
		err = errors.Join(err, fmt.Errorf("runtime name is required"))
	} else if !types.TemplateVarNamePattern.MatchString(r.Name) {
		err = errors.Join(err, fmt.Errorf("invalid runtime name: %s", r.Name))
	}

	return err
}

func (r *Runtime) Clone() *Runtime {
	if r == nil {
		return nil
	}
	return &Runtime{
		Type: r.Type,
		Name: r.Name,
	}
}
