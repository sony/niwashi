// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import (
	"github.com/sony/niwashi/internal/profile"
)

type Runtime struct {
	Bindings  map[string]string `json:"capabilityBinding" yaml:"capabilityBinding"`
	ToolAlias map[string]string `json:"toolAlias" yaml:"toolAlias"`
	Params    *profile.Params   `json:"params" yaml:"params"`
}

func (r *Runtime) SetDefaults() {
	if r.Bindings == nil {
		r.Bindings = make(map[string]string)
	}
	if r.ToolAlias == nil {
		r.ToolAlias = make(map[string]string)
	}
}
