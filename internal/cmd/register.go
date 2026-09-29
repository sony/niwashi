// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmd

import (

	// Action implementations
	_ "github.com/sony/niwashi/internal/action/exec_local"
	_ "github.com/sony/niwashi/internal/action/exec_remote"

	// Transport implementations
	_ "github.com/sony/niwashi/internal/transport/ssh"
	_ "github.com/sony/niwashi/internal/transport/winrm"

	// Phase implementations
	_ "github.com/sony/niwashi/internal/workflow/phase/barrier"
	_ "github.com/sony/niwashi/internal/workflow/phase/cluster"
	_ "github.com/sony/niwashi/internal/workflow/phase/cluster_sync"
	_ "github.com/sony/niwashi/internal/workflow/phase/host"
	_ "github.com/sony/niwashi/internal/workflow/phase/infra"
	_ "github.com/sony/niwashi/internal/workflow/phase/node"
	_ "github.com/sony/niwashi/internal/workflow/phase/node_sync"
)
