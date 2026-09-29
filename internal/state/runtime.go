// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/clone"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/types"
)

type Runtime struct {
	ToolAliases       map[string]string `json:"toolAliases,omitempty" yaml:"toolAliases,omitempty"`
	CapabilityBinding map[string]string `json:"capabilityBinding,omitempty" yaml:"capabilityBinding,omitempty"`
	Tool              types.Dict        `json:"tool,omitempty" yaml:"tool,omitempty"`
	Service           types.Dict        `json:"service,omitempty" yaml:"service,omitempty"` // for runtime use, not merged or applied
}

func (r *Runtime) GetToolAlias() map[string]string {
	return r.ToolAliases
}

func (r *Runtime) GetCapabilityBinding() map[string]string {
	return r.CapabilityBinding
}

func (r *Runtime) GetTools() types.Dict {
	return r.Tool
}

func (r *Runtime) GetService() types.Dict {
	return r.Service
}

func (r *Runtime) Clone() *Runtime {
	if r == nil {
		return nil
	}

	cloned := &Runtime{}
	cloned.ToolAliases = clone.StringMap(r.ToolAliases)
	cloned.CapabilityBinding = clone.StringMap(r.CapabilityBinding)
	cloned.Tool = r.Tool.Clone()
	cloned.Service = r.Service.Clone()

	return cloned
}

func (r *Runtime) SetDefaults() {
	if r.ToolAliases == nil {
		r.ToolAliases = make(map[string]string)
	}
	if r.CapabilityBinding == nil {
		r.CapabilityBinding = make(map[string]string)
	}
	if r.Tool == nil {
		r.Tool = types.Dict{}
	}
	if r.Service == nil {
		r.Service = types.Dict{}
	}
}

func (r *Runtime) Validate() error {
	var err error

	if r.Tool != nil {
		if e := r.Tool.ValidateAsTemplateVar(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid runtime tool: %w", e))
		}
	}
	if r.Service != nil {
		if e := r.Service.ValidateAsTemplateVar(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid runtime service: %w", e))
		}
	}

	return err
}

func (r *Runtime) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	// only "tool" is supported
	switch path[0] {
	case "tool":
		return r.Tool.SetPath(value, path[1:]...)
	case "service":
		return r.Service.SetPath(value, path[1:]...)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (r *Runtime) Remove(path []string) error {
	return fmt.Errorf("Remove is not implemented yet")
}

func (*Runtime) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Runtime")
}
