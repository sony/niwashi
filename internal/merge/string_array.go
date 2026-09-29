// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package merge

import "slices"

func StringArray(a, b []string) []string {
	// Don't merge arrays, just override
	if len(b) == 0 {
		return slices.Clone(a)
	}
	return slices.Clone(b)
}
