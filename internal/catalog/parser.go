// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package catalog

import (
	"io"

	"github.com/sony/niwashi/internal/logger"
	"gopkg.in/yaml.v3"
)

func Parse(r io.Reader) *Catalog {

	data, err := io.ReadAll(r)
	if err != nil {
		logger.Error("Read error:", err)
		return nil
	}

	catalog := &Catalog{}
	err = yaml.Unmarshal(data, catalog)
	if err != nil {
		logger.Error("Unmarshal error:", err)
		return nil
	}

	return catalog
}
