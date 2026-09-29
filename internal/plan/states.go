// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import "github.com/sony/niwashi/internal/state"

type States struct {
	Initial InitState    `json:"initial" yaml:"initial"`
	Target  *state.State `json:"target" yaml:"target"`
}
type InitState struct {
	Hash string `json:"hash" yaml:"hash"`
}

func (s *States) SetDefaults() {
	if s.Target == nil {
		s.Target = state.NewState()
	}
	s.Target.SetDefaults()
}
