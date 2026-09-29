// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workspace"
	"github.com/spf13/cobra"
)

type ExportData struct {
	WorkDir        string
	Out            string
	IncludeRuntime bool
}

var exportData ExportData

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the current state to a file",
		Long: `Exports the current infrastructure state from the workspace
into a YAML file. This snapshot can be used as an initial state for
future plans, enabling easy replication or migration of configurations.`,
		SilenceUsage: true, // Don't show usage on error
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunExportCmd(&exportData)
		},
	}

	cmd.Flags().StringVar(&exportData.WorkDir, FlagWorkDir, DefaultWorkDir, UsageWorkDir)
	cmd.Flags().StringVarP(&exportData.Out, "out", "o", "", "Path to the output YAML file (default: stdout)")
	cmd.Flags().BoolVar(&exportData.IncludeRuntime, "include-runtime", false, "Include the runtime section in the export")

	return cmd
}

func RunExportCmd(arg *ExportData) (err error) {

	fs := file.NewDefaultFileSystem()
	ws := workspace.NewRunWs(arg.WorkDir)
	reader, err := fs.Open(ws.GetStateFilePath())
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()

	s, err := state.LoadAsJson(reader)
	if err != nil {
		return err
	}

	var output io.Writer
	if arg.Out == "" {
		output = os.Stdout
	} else {
		file, err := fs.OpenFile(arg.Out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, file.Close()) }()
		output = file
	}

	// Exclude runtime section if not requested
	if !arg.IncludeRuntime {
		s.Runtime = nil
	}

	err = s.WriteAsYaml(output)
	if err != nil {
		return err
	}

	fmt.Println("Export completed successfully.")

	return nil
}
