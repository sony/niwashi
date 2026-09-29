// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package cmd

import (
	"fmt"

	"github.com/sony/niwashi/internal/system"
	"github.com/spf13/cobra"
)

// sshCmd represents the ssh command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "version",
		Short:        "Show the version of nwsctl",
		Long:         `Show the version of nwsctl.`,
		SilenceUsage: true, // Don't show usage on error
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersionCmd()
		},
	}

	return cmd
}

func RunVersionCmd() error {
	fmt.Println("nwsctl version", system.CliVersion.Version)
	fmt.Println("Build information:")
	fmt.Println("* Go version:", system.CliVersion.GoVersion)
	fmt.Println("* Git revision:", system.CliVersion.Commit)
	fmt.Println("* Build date:", system.CliVersion.Date)
	return nil
}
