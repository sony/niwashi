// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package catalog

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/file"
)

func Load(p string, opener file.Opener) (*Catalog, error) {
	var catalog Catalog
	err := file.LoadWithDecoding(&catalog, p, nil, opener)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to load catalog from: %q", p))
	}
	return &catalog, nil
}
