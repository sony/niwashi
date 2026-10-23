// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import (
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/tmpl"
	"github.com/sony/niwashi/internal/types"
)

type Template struct {
	From        string        `json:"from,omitempty" yaml:"from,omitempty"`
	Provisioner string        `json:"provisioner,omitempty" yaml:"provisioner,omitempty"`
	Params      types.Params  `json:"params,omitempty" yaml:"params,omitempty"`
	Import      types.Imports `json:"import,omitempty" yaml:"import,omitempty"`
}

func (it *Template) Clone() *Template {
	if it == nil {
		return nil
	}
	clone := &Template{}
	clone.From = it.From
	clone.Provisioner = it.Provisioner
	clone.Params = it.Params.Clone()
	clone.Import = it.Import.Clone()
	return clone
}

func (it *Template) Validate() error {
	if err := it.Params.ValidateAsParams(); err != nil {
		return err
	}
	if err := it.Import.Validate(); err != nil {
		return err
	}

	return nil
}

func (it *Template) Merge(other *Template) (*Template, error) {
	if it == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return it.Clone(), nil
	}

	clone := &Template{}
	var err error
	clone.From = merge.String(it.From, other.From)
	clone.Provisioner = merge.String(it.Provisioner, other.Provisioner)
	clone.Params, err = types.Merge(it.Params, other.Params)
	if err != nil {
		return nil, err
	}
	clone.Import = it.Import.Merge(other.Import)
	return clone, nil
}

func (it *Template) FromField() string {
	return it.From
}

func (it *Template) Resolve(tpl map[string]*Template) (*Template, error) {
	merged, err := tmpl.Resolve(tpl, it)
	if err != nil {
		return nil, err
	}

	// remove From field after merge
	merged.From = ""

	return merged, nil
}
