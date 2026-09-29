// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package profile

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/file"
)

func Load(p string, opener file.Opener) (*Profile, error) {
	var profile Profile
	err := file.LoadWithDecoding(&profile, p, nil, opener)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to load profile from: %q", p))
	}

	// todo: validate

	profile.SetDefaults()

	return &profile, nil
}
