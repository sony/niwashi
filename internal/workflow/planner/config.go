// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import (
	"fmt"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/system"
	"github.com/sony/niwashi/internal/workspace"
)

type Config struct {
	fs          file.FileSystem
	mode        plan.Mode
	WorkDir     string
	profilePath string
	RecipeDirs  []string
	TargetFiles []string
}

func NewConfig(opts ...func(*Config)) *Config {
	cfg := &Config{
		mode: plan.ModeDefault,
		fs:   file.NewDefaultFileSystem(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func (cfg *Config) Validate() error {

	switch cfg.mode {
	case plan.ModeDefault:
		// require at least one target file for default mode
		if len(cfg.TargetFiles) == 0 {
			return fmt.Errorf("at least one target file is required")
		}
	case plan.ModeDestroy:
		if len(cfg.TargetFiles) > 0 {
			return fmt.Errorf("target files should not be specified for mode: %s", cfg.mode)
		}
	case plan.ModePrune:
		if len(cfg.TargetFiles) == 0 {
			return fmt.Errorf("at least one target file is required")
		}
	default:
		return fmt.Errorf("invalid mode: %s", cfg.mode)
	}

	return nil
}

func WithFileSystem(fs file.FileSystem) func(*Config) {
	return func(cfg *Config) {
		cfg.fs = fs
	}
}

func WithWorkDir(workDir string) func(*Config) {
	return func(cfg *Config) {
		cfg.WorkDir = workDir
	}
}

func WithProfile(profilePath string) func(*Config) {
	return func(cfg *Config) {
		cfg.profilePath = profilePath
	}
}

func WithRecipeDirs(dirs ...string) func(*Config) {
	return func(cfg *Config) {
		cfg.RecipeDirs = dirs
	}
}

func WithTargetFiles(paths ...string) func(*Config) {
	return func(cfg *Config) {
		cfg.TargetFiles = paths
	}
}

func WithDestroy() func(*Config) {
	return func(cfg *Config) {
		cfg.mode = plan.ModeDestroy
	}
}

func WithPrune() func(*Config) {
	return func(cfg *Config) {
		cfg.mode = plan.ModePrune
	}
}

func (cfg *Config) GetMode() plan.Mode {
	return cfg.mode
}

func (cfg *Config) LoadInitialState() (*state.State, error) {

	ws := workspace.NewRunWs(cfg.WorkDir)

	logger.Debug("Loading state file:", "workdir", cfg.WorkDir)
	f, err := cfg.fs.Open(ws.GetStateFilePath())
	if err != nil {
		logger.Error("failed to open initial state file:", "error", err)
		return nil, err
	}
	defer func() { _ = f.Close() }()

	initial, err := state.LoadAsJson(f)
	if err != nil {
		logger.Error("failed to load initial state:", "error", err)
		return nil, err
	}
	initial.SetDefaults()
	return initial, nil
}

func (cfg *Config) LoadTargetState() (*state.State, error) {
	if len(cfg.TargetFiles) == 0 {
		return state.NewState(), nil
	}

	var target *state.State
	for _, p := range cfg.TargetFiles {
		logger.Debug("Loading state file:", "path", p)
		state, err := state.Load(p, cfg.fs)
		if err != nil {
			logger.Error("failed to load target state:", "error", err)
			return nil, err
		}

		if target == nil {
			target = state
		} else {
			target, err = target.Merge(state)
			if err != nil {
				logger.Error("failed to merge target state:", "error", err)
				return nil, err
			}
		}
	}
	target.SetDefaults()
	return target, nil
}

func (cfg *Config) LoadProfile() (*profile.Profile, error) {
	if cfg.profilePath != "" {
		prof, err := profile.Load(cfg.profilePath, cfg.fs)
		if err != nil {
			return nil, err
		}
		prof.SetDefaults()
		return prof, nil
	} else {
		return profile.NewProfile(), nil
	}
}

func (cfg *Config) LoadRecipeFinder() (cmap.RecipeFinder, error) {
	return cmap.NewCapabilityMapFrom(
		system.NewSystemRecipeLoader(),
		cfg.fs,
		cfg.RecipeDirs...)
}
