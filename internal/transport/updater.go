// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import "github.com/sony/niwashi/internal/workspace"

type Updater interface {
	Update(t Transport, ws *workspace.RunWorkspace) ([]Patch, error)
}
