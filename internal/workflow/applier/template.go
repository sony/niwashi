// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"github.com/sony/niwashi/internal/action"
)

func RenderTemplate(input string, val any) (string, error) {
	return action.NewTemplate(action.NewFuncMap(val)).Render(input, val)
}
