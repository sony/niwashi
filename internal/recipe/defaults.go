// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/clone"
	"github.com/sony/niwashi/internal/types"
)

type Defaults struct {
	Params       types.Params      `json:"params" yaml:"params"`
	Environments map[string]string `json:"env" yaml:"env"`
}

func (d *Defaults) SetDefaults() {
	if d.Params == nil {
		d.Params = make(types.Params)
	}
	if d.Environments == nil {
		d.Environments = make(map[string]string)
	}
}

func (d *Defaults) Validate() error {
	err := d.Params.ValidateAsParams()
	if err != nil {
		err = errors.Join(err, fmt.Errorf("invalid params: %w", err))
	}

	for k := range d.Environments {
		if !types.EnvVarNamePattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid env name: %s", k))
		}
	}

	return err
}

func (d *Defaults) Clone() *Defaults {
	if d == nil {
		return nil
	}
	return &Defaults{
		Params:       d.Params.Clone(),
		Environments: clone.StringMap(d.Environments),
	}
}
