// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package merge

import "maps"

func StringMap(a, b map[string]string) map[string]string {
	merged := maps.Clone(a)
	if merged == nil {
		merged = make(map[string]string)
	}

	for k, v := range b {
		// Skip empty value in b
		if len(v) == 0 {
			continue
		}
		merged[k] = v
	}
	return merged
}
