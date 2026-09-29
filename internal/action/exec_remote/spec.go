// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/types"
)

type Spec struct {
	ScriptTemplate string            `json:"scriptTpl" yaml:"scriptTpl"`
	EnvTemplate    map[string]string `json:"envTpl" yaml:"envTpl"`
}

func (s *Spec) GetEnvTpl() map[string]string {
	return s.EnvTemplate
}

func (s *Spec) Validate() error {
	var err error

	for k := range s.EnvTemplate {
		if !types.EnvVarNamePattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid env name: %s", k))
		}
	}

	return err
}
