// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import "github.com/sony/niwashi/internal/merge"

type Selector struct {
	Generator string `json:"generator,omitempty" yaml:"generator,omitempty"`
	Instance  string `json:"instance,omitempty" yaml:"instance,omitempty"`
}

func (is *Selector) Clone() *Selector {
	if is == nil {
		return nil
	}

	return &Selector{
		Generator: is.Generator,
		Instance:  is.Instance,
	}
}

func (is *Selector) Merge(other *Selector) *Selector {
	if is == nil {
		return other.Clone()
	}
	if other == nil {
		return is.Clone()
	}

	merged := &Selector{
		Generator: merge.String(is.Generator, other.Generator),
		Instance:  merge.String(is.Instance, other.Instance),
	}

	return merged
}
