// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package system

import "time"

type VersionData struct {
	Version   string
	GoVersion string
	Commit    string
	Date      string
}

var (
	CliVersion VersionData = VersionData{
		Version:   "unknown",
		GoVersion: "unknown",
		Commit:    "unknown",
		Date:      "unknown",
	}

	LaunchAt time.Time = time.Now()
)
