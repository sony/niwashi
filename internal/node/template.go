// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package node

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/clone"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/types"
)

type Template struct {
	From         string             `json:"from,omitempty" yaml:"from,omitempty"`
	Os           string             `json:"os,omitempty" yaml:"os,omitempty"`
	Labels       map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty"`
	Capabilities cap.CapabilityList `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

func (n *Template) Clone() *Template {
	if n == nil {
		return nil
	}
	clone := &Template{
		From:         n.From,
		Os:           n.Os,
		Labels:       clone.StringMap(n.Labels),
		Capabilities: n.Capabilities.Clone(),
	}
	return clone
}

func (nt *Template) Validate() error {
	var err error

	if e := nt.Capabilities.Validate(); e != nil {
		err = errors.Join(err, e)
	}

	for k := range nt.Labels {
		if !types.LabelKeyPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid label key: %q", k))
		}
	}

	return err
}

func (nt *Template) Merge(other *Template) (*Template, error) {
	if nt == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return nt.Clone(), nil
	}
	clone := &Template{}
	var err error
	clone.From = merge.String(nt.From, other.From)
	clone.Os = merge.String(nt.Os, other.Os)
	clone.Labels = merge.StringMap(nt.Labels, other.Labels)
	clone.Capabilities, err = nt.Capabilities.Merge(other.Capabilities)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (n *Template) FromField() string {
	return n.From
}
