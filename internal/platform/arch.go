// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package platform

// Based on GOARCH values: https://golang.org/doc/install/source#environment
const (
	ArchUnknown string = "unknown"
	ArchAMD64   string = "amd64"
	ArchARM64   string = "arm64"
	ArchARM     string = "arm"
	Arch386     string = "386"
)

var archAliases = map[string]string{
	"amd64":   ArchAMD64,
	"AMD64":   ArchAMD64,
	"x86_64":  ArchAMD64,
	"X86_64":  ArchAMD64,
	"arm64":   ArchARM64,
	"ARM64":   ArchARM64,
	"aarch64": ArchARM64,
	"AARCH64": ArchARM64,
	"arm":     ArchARM,
	"ARM":     ArchARM,
	"386":     Arch386,
	"i386":    Arch386,
	"I386":    Arch386,
}

func NormalizeArch(arch string) string {
	if normalized, ok := archAliases[arch]; ok {
		return normalized
	}
	return arch
}
