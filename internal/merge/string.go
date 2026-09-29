// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package merge

func String(a, b string) string {
	if len(b) == 0 {
		return a
	}
	return b
}
