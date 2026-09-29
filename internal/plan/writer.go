// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import (
	"github.com/sony/niwashi/internal/file"
)

func Write(p *Plan, path string, creator file.Creator) error {
	return file.WriteWithEncoding(p, path, creator)
}
