// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package catalog

const Version = "nws.catalog/v1"

type Catalog struct {
	Version string            `json:"version" yaml:"version"`
	Recipes map[string]Recipe `json:"recipes" yaml:"recipes"`
}

type Recipe struct {
	Path string `json:"path" yaml:"path"`
}

func (s *Catalog) GetVersion() string {
	return s.Version
}
