// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"testing"
)

func TestDefaultShell(t *testing.T) {
	tests := []struct {
		os   string
		want string
	}{
		// Windows PowerShell ("powershell"), not PowerShell 7+ ("pwsh"):
		// pwsh requires a separate install and isn't present on a stock
		// Windows image, while Windows PowerShell ships with every
		// supported Windows version.
		{os: "windows", want: "powershell"},
		{os: "linux", want: "/bin/sh"},
		{os: "unknown", want: "/bin/sh"},
		{os: "", want: "/bin/sh"},
	}
	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			if got := DefaultShell(tt.os); got != tt.want {
				t.Errorf("DefaultShell(%q) = %q, want %q", tt.os, got, tt.want)
			}
		})
	}
}

func TestGetExtension(t *testing.T) {
	tests := []struct {
		shebang string
		os      string
		want    string
	}{
		{shebang: "", os: "windows", want: ".ps1"},
		{shebang: "", os: "linux", want: ".sh"},
		{shebang: "#!/bin/bash", os: "linux", want: ".sh"},
		{shebang: "#!/bin/bash", os: "windows", want: ".sh"},
		{shebang: "#!/bin/sh", os: "linux", want: ".sh"},
		{shebang: "#!powershell", os: "windows", want: ".ps1"},
		{shebang: "#!powershell", os: "linux", want: ".ps1"},
		{shebang: "#!/usr/bin/env powershell", os: "linux", want: ".ps1"},
		{shebang: "#!/usr/bin/env powershell", os: "windows", want: ".ps1"},
		{shebang: "#!/usr/local/bin/pwsh", os: "linux", want: ".ps1"},
		{shebang: "#!/usr/local/bin/pwsh", os: "windows", want: ".ps1"},
	}
	for _, tt := range tests {
		t.Run(tt.shebang, func(t *testing.T) {
			if got := GetExtension(tt.shebang, tt.os); got != tt.want {
				t.Errorf("GetExtension(%q, %q) = %v, want %v", tt.shebang, tt.os, got, tt.want)
			}
		})
	}
}

func TestScript_GetArgs_InsertsFileFlagAfterShebangArgs(t *testing.T) {
	// linux default shell = /bin/sh
	// windows default shell = powershell

	tests := []struct {
		name   string
		script Script
		os     string
		want   []string
	}{
		{
			name: "powershell script with shebang",
			script: Script{
				firstLine: "#!powershell",
				localPath: "C:\\script.ps1",
			},
			os:   "windows",
			want: []string{"powershell", "-File", "C:\\script.ps1"},
		},
		{
			name: "powershell script with shebang and -NoProfile",
			script: Script{
				firstLine: "#!powershell -NoProfile",
				localPath: "C:\\script.ps1",
			},
			os:   "windows",
			want: []string{"powershell", "-NoProfile", "-File", "C:\\script.ps1"},
		},
		{
			name: "powershell script without shebang",
			script: Script{
				firstLine: "",
				localPath: "C:\\script.ps1",
			},
			os:   "windows",
			want: []string{"powershell", "-File", "C:\\script.ps1"},
		},
		{
			name: "bash script with shebang on linux",
			script: Script{
				firstLine: "#!/bin/bash",
				localPath: "/tmp/script.sh",
			},
			os:   "linux",
			want: []string{"/bin/bash", "/tmp/script.sh"},
		},
		{
			name: "bash script with shebang+/usr/bin/env on linux",
			script: Script{
				firstLine: "#!/usr/bin/env bash",
				localPath: "/tmp/script.sh",
			},
			os:   "linux",
			want: []string{"/usr/bin/env", "bash", "/tmp/script.sh"},
		},
		{
			name: "sh script without shebang on linux",
			script: Script{
				firstLine: "",
				localPath: "/tmp/script.sh",
			},
			os:   "linux",
			want: []string{"/bin/sh", "/tmp/script.sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.script.GetArgs(tt.os)
			if !equalSlices(args, tt.want) {
				t.Errorf("GetArgs(%q) = %v, want %v", tt.os, args, tt.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
