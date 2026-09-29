// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package platform

// Based on GOOS values: https://golang.org/doc/install/source#environment
const (
	OsUnknown string = "unknown"
	OsLinux   string = "linux"
	OsDarwin  string = "darwin"
	OsWindows string = "windows"
)

var osAliases = map[string]string{
	"linux":   OsLinux,
	"Linux":   OsLinux,
	"darwin":  OsDarwin,
	"Darwin":  OsDarwin,
	"windows": OsWindows,
	"Windows": OsWindows,
}

func NormalizeOs(os string) string {
	if normalized, ok := osAliases[os]; ok {
		return normalized
	}
	return os
}
