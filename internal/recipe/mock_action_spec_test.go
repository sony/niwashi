// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

type mockActionSpec struct{}

func (s *mockActionSpec) GetEnvTpl() map[string]string {
	return nil
}

func (s *mockActionSpec) Validate() error {
	return nil
}
