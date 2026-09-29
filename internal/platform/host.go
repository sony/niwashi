// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package platform

import "runtime"

const (
	HostOs   string = runtime.GOOS
	HostArch string = runtime.GOARCH
)
