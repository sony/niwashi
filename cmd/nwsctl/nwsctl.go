// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package main

import (
	"runtime/debug"

	"github.com/sony/niwashi/internal/cmd"
	"github.com/sony/niwashi/internal/system"
)

var (
	version = "unknown"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	system.CliVersion = makeVersionData()

	cmd.Execute()
}

func makeVersionData() system.VersionData {
	goVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		goVersion = info.GoVersion
	}

	return system.VersionData{
		Version:   version,
		GoVersion: goVersion,
		Commit:    commit,
		Date:      date,
	}
}
