// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package clone

import "strings"

func StringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}

	clone := make(map[string]string, len(m))
	for k, v := range m {
		clone[k] = strings.Clone(v)
	}
	return clone
}
