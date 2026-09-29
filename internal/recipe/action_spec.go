// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/logger"
)

func NewActionSpec(t string) (action.ActionSpec, error) {
	return action.NewSpec(t)
}

type NopLoader struct{}

func (l *NopLoader) NewSpec() action.ActionSpec {
	return &nopSpec{}
}

func (l *NopLoader) Build(ctx action.ActionContext) (action.Action, error) {
	return nil, nil
}

type nopSpec struct{}

func (s *nopSpec) GetEnvTpl() map[string]string {
	return nil
}

func (s *nopSpec) Validate() error {
	return nil
}

func EnableNopLoader() {
	logger.Info("Enabling NopLoader for ActionSpec")
	action.RegisterBuilder("nop", &NopLoader{})
}
