// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package profile

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/types"
)

const Version = "nws.profile/v1"

type Profile struct {
	Version           string            `json:"version" yaml:"version"`
	Name              string            `json:"name" yaml:"name"`
	ToolAlias         map[string]string `json:"toolAlias" yaml:"toolAlias"`
	CapabilityBinding map[string]string `json:"capabilityBinding" yaml:"capabilityBinding"`
	Params            *Params           `json:"params" yaml:"params"`
}

type Params struct {
	Clusters     map[string]types.Params `json:"cluster" yaml:"cluster"`
	Capabilities map[string]types.Params `json:"capability" yaml:"capability"`
}

func NewProfile() *Profile {
	p := &Profile{
		Version: Version,
	}
	p.SetDefaults()
	return p
}

func (p *Profile) SetDefaults() {
	if p.ToolAlias == nil {
		p.ToolAlias = map[string]string{}
	}
	if p.CapabilityBinding == nil {
		p.CapabilityBinding = map[string]string{}
	}
	if p.Params == nil {
		p.Params = &Params{}
	}
	if p.Params.Clusters == nil {
		p.Params.Clusters = map[string]types.Params{}
	}
	if p.Params.Capabilities == nil {
		p.Params.Capabilities = map[string]types.Params{}
	}
}

func (p *Profile) Validate() error {
	var err error
	if p.Params != nil {
		for k, v := range p.Params.Capabilities {
			e := v.ValidateAsParams()
			if e != nil {
				err = errors.Join(err, fmt.Errorf("invalid capability params for %s: %w", k, e))
			}
		}

		for k, v := range p.Params.Clusters {
			e := v.ValidateAsParams()
			if e != nil {
				err = errors.Join(err, fmt.Errorf("invalid cluster params for %s: %w", k, e))
			}
		}
	}

	return err
}
