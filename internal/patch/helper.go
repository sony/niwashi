// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

import (
	"fmt"
	"strings"
)

func ParsePath(path string) ([]string, error) {

	if path == "" || path[0] != '/' {
		return nil, fmt.Errorf("invalid patch path: %s", path)
	}
	parts := strings.Split(path[1:], "/")

	// unescape ~1 to /
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(p, "~1", "/")
	}
	return parts, nil
}
