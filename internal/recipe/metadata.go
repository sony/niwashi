// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import "fmt"

type Metadata struct {
	Id          string `json:"id" yaml:"id"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

func (m *Metadata) Validate() error {
	if m.Id == "" {
		return fmt.Errorf("metadata.id is required")
	} else if !recipeIdPattern.MatchString(m.Id) {
		return fmt.Errorf("metadata.id must be in the format org/name: %q", m.Id)
	}

	if m.Version == "" {
		return fmt.Errorf("metadata.version is required")
	} else if !versionPattern.MatchString(m.Version) {
		return fmt.Errorf("metadata.version must be in the format X.Y.Z where X, Y, and Z are positive integers: %s", m.Version)
	}

	return nil
}
