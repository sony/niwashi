// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

type recipeSpec struct {
	fqid          string
	dir           string
	assets        []string
	defaultParams map[string]any
	defaultEnv    map[string]string
}

func (s *recipeSpec) GetFqid() string {
	return s.fqid
}

func (s *recipeSpec) GetDir() string {
	return s.dir
}

func (s *recipeSpec) GetAssets() []string {
	return s.assets
}

func (s *recipeSpec) GetDefaultParams() map[string]any {
	return s.defaultParams
}

func (s *recipeSpec) GetDefaultEnv() map[string]string {
	return s.defaultEnv
}
