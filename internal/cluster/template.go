// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster

import (
	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/types"
)

type Template struct {
	From         string             `json:"from,omitempty" yaml:"from,omitempty"`
	Version      string             `json:"version,omitempty" yaml:"version,omitempty"`
	Params       types.Params       `json:"params,omitempty" yaml:"params,omitempty"`
	Capabilities cap.CapabilityList `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

func (ct *Template) Clone() *Template {
	if ct == nil {
		return nil
	}
	clone := &Template{}
	clone.From = ct.From
	clone.Version = ct.Version
	clone.Params = ct.Params.Clone()
	clone.Capabilities = ct.Capabilities.Clone()
	return clone
}

func (ct *Template) Validate() error {
	err := ct.Params.ValidateAsParams()
	if err != nil {
		return err
	}

	err = ct.Capabilities.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (ct *Template) Merge(other *Template) (*Template, error) {
	if ct == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return ct.Clone(), nil
	}

	clone := &Template{}
	var err error
	clone.From = merge.String(ct.From, other.From)
	clone.Version = merge.String(ct.Version, other.Version)
	clone.Params, err = types.Merge(ct.Params, other.Params)
	if err != nil {
		return nil, err
	}
	clone.Capabilities, err = ct.Capabilities.Merge(other.Capabilities)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (ct *Template) FromField() string {
	return ct.From
}
