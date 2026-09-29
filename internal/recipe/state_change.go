// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"fmt"
	"slices"
	"strings"

	"github.com/sony/niwashi/internal/logger"
)

const (
	OpSet    = "set"
	OpAdd    = "add" // deprecated, use "set" instead
	OpRemove = "remove"
)

var (
	patchOperations = []string{OpSet, OpRemove, OpAdd}
)

type StateChangeOperation struct {
	Op            string `json:"op" yaml:"op"`
	Path          string `json:"path" yaml:"path"`
	Value         any    `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFromFile string `json:"valueFromFile,omitempty" yaml:"valueFromFile,omitempty"`
	ValueFromJson string `json:"valueFromJson,omitempty" yaml:"valueFromJson,omitempty"`
	Count         string `json:"count,omitempty" yaml:"count,omitempty"`
}

func (s *StateChangeOperation) Validate() error {
	if !slices.Contains(patchOperations, s.Op) {
		return fmt.Errorf("invalid state change operation %q, must be one of %s", s.Op, strings.Join(patchOperations, ", "))
	}

	// show deprecated warning for "add" operation
	if s.Op == OpAdd {
		logger.Warn("the 'add' operation is deprecated, please use 'set' instead")
	}

	switch s.Op {
	case OpSet, OpAdd:
		if s.Value == nil && s.ValueFromFile == "" && s.ValueFromJson == "" {
			return fmt.Errorf("value, valueFromFile, or valueFromJson is required for add operation")
		}
	case OpRemove:
		if s.Value != nil || s.ValueFromFile != "" || s.ValueFromJson != "" {
			logger.Warn("value, valueFromFile, and valueFromJson are ignored for remove operation")
		}
	}

	return nil
}
