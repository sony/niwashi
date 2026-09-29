// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import "strings"

var (
	charsToReplace = []string{
		"@", "_",
		"/", "_",
		"\\", "_",
		" ", "_",
	}
)

func SanitizeFilename(str string) string {

	replacer := strings.NewReplacer(charsToReplace...)
	return replacer.Replace(str)
}
