// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workspace"
	"github.com/spf13/cobra"
)

type InitData struct {
	WorkDir    string
	StateFile  string
	filesystem file.FileSystem
}

var initData InitData

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the workspace",
		Long: `Initializes the workspace for nwsctl by creating necessary
		files and directories.
		If no state file is specified, an empty state file will be created in the workspace.`,
		SilenceUsage: true, // Don't show usage on error
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunInitCmd(&initData)
		},
	}

	cmd.Flags().StringVar(&initData.WorkDir, FlagWorkDir, DefaultWorkDir, UsageWorkDir)
	cmd.Flags().StringVarP(&initData.StateFile, "state", "s", "", "Path to the state YAML file")

	return cmd
}

func RunInitCmd(arg *InitData) error {

	filesystem := arg.filesystem
	if filesystem == nil {
		filesystem = file.NewDefaultFileSystem()
	}

	ws := workspace.NewRootWorkspace(arg.WorkDir)

	stateFilePath := ws.GetStateFilePath()

	// if file already exists, returns error
	_, err := filesystem.Stat(stateFilePath)
	if err == nil {
		logger.Error("state file already exists")
		return fmt.Errorf("state file already exists: %s", stateFilePath)
	}

	// Create directory and save state
	stateDir := filepath.Dir(stateFilePath)
	if err := file.MakeDirsAll(filesystem, stateDir); err != nil {
		logger.Error("failed to create workspace directories:", "path", stateDir, "error", err)
		return err
	}

	// load or create state
	var s *state.State
	if arg.StateFile == "" {
		s = state.NewState()
	} else {

		// Load state file once, and save state as JSON format
		ext := filepath.Ext(arg.StateFile)
		var loader func(io.ReadSeeker) (*state.State, error)
		switch ext {
		case ".yaml", ".yml":
			loader = state.LoadAsYaml
		case ".json", ".jsonc":
			loader = state.LoadAsJson
		default:
			return fmt.Errorf("unsupported state file format: %s", arg.StateFile)
		}

		f, err := filesystem.Open(arg.StateFile)
		if err != nil {
			logger.Error("failed to open state file:", "path", arg.StateFile, "error", err)
			return err
		}
		defer func() { _ = f.Close() }()

		s, err = loader(f)
		if err != nil {
			logger.Error("failed to load state file:", "path", arg.StateFile, "error", err)
			return err
		}
	}

	w, err := filesystem.OpenFile(stateFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	err = s.WriteAsJson(w)
	return errors.Join(err, w.Close())
}
