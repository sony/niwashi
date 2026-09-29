// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import (
	"fmt"

	"github.com/sony/niwashi/internal/file"
)

func Load(p string, opener file.Opener) (*Plan, error) {
	var plan Plan
	var hash string
	err := file.LoadWithDecoding(&plan, p, &hash, opener)
	if err != nil {
		return nil, fmt.Errorf("fail to load plan from: %q: %w", p, err)
	}

	plan.FilePath = p
	plan.Hash = hash

	plan.SetDefaults()

	return &plan, nil
}
