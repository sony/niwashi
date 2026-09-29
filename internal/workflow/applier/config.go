// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/system"
	"github.com/sony/niwashi/internal/workspace"
)

const DefaultConcurrency uint = 5

type Config struct {
	RecipeDirs  []string
	PlanFile    string
	WorkDir     string
	DryRunMode  string
	Timestamp   time.Time
	DevOption   *DevOption
	FileSystem  file.FileSystem
	Mode        plan.Mode
	Concurrency uint
	StrictCheck bool
}

func NewConfig() *Config {
	return &Config{
		Mode:        plan.ModeDefault,
		FileSystem:  file.NewDefaultFileSystem(),
		Timestamp:   time.Now().UTC(),
		Concurrency: DefaultConcurrency,
	}
}

func (cfg *Config) Validate() error {
	var err error

	if cfg.WorkDir == "" {
		err = errors.Join(err, fmt.Errorf("valid --work-dir path is required"))
	}

	for _, dir := range cfg.RecipeDirs {
		if dir == "" {
			err = errors.Join(err, fmt.Errorf("--recipe-dir contains empty string"))
		} else {
			// check if the directory exists
			if _, statErr := cfg.FileSystem.Stat(dir); statErr != nil {
				if errors.Is(statErr, fs.ErrNotExist) {
					err = errors.Join(err, fmt.Errorf("--recipe-dir does not exist: %q", dir))
				} else {
					err = errors.Join(err, fmt.Errorf("fail to access --recipe-dir: %q", dir))
				}
			}
		}
	}

	if cfg.PlanFile == "" {
		err = errors.Join(err, fmt.Errorf("valid --plan path is required"))
	} else {
		// check if the plan file exists
		if _, statErr := cfg.FileSystem.Stat(cfg.PlanFile); statErr != nil {
			if errors.Is(statErr, fs.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("--plan does not exist: %q", cfg.PlanFile))
			} else {
				err = errors.Join(err, fmt.Errorf("fail to access --plan file: %q", cfg.PlanFile))
			}
		}
	}

	return err
}

func (cfg *Config) LoadPlan() (*plan.Plan, error) {
	return plan.Load(cfg.PlanFile, cfg.FileSystem)
}

func (cfg *Config) NewWorkspace(planHash string) (*workspace.RunWorkspace, error) {

	root, err := filepath.Abs(cfg.WorkDir)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to get absolute path of workDir: %q", cfg.WorkDir))
	}

	// Create DryRun root path
	var runsWs *workspace.RootWorkspace
	realWs := workspace.NewRootWorkspace(root)

	if cfg.DryRunMode == "" || cfg.DryRunMode == "off" {
		runsWs = realWs
	} else {
		dryRunRoot := realWs.GetDryRunRootPath()

		runsWs = workspace.NewRootWorkspace(dryRunRoot)
		err := cfg.setupSimData(runsWs, realWs)
		if err != nil {
			return nil, errors.Join(err, fmt.Errorf("fail to setup data for dry-run"))
		}
	}

	runId, err := cfg.getRunsId(runsWs, planHash)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to get run ID"))
	}

	return workspace.NewRunWorkspace(runsWs, runId), nil
}

func (cfg *Config) setupSimData(simEnv, realEnv *workspace.RootWorkspace) (err error) {

	// create dry-run dir first
	err = file.MakeDirsAll(
		cfg.FileSystem,
		simEnv.GetRootDirPath(),
		simEnv.GetStateRootPath(),
		// other directories will be created by applier later.
	)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to create dry-run directory"))
	}

	// copy state file to dry-run directory
	src, err := cfg.FileSystem.Open(realEnv.GetStateFilePath())
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to open state file from real environment"))
	}
	defer func() { _ = src.Close() }()

	dest, err := cfg.FileSystem.OpenFile(
		simEnv.GetStateFilePath(),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to open state file in dry-run directory"))
	}
	defer func() {
		err = errors.Join(err, dest.Close())
	}()

	_, err = io.Copy(dest, src)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to copy state file to dry-run directory"))
	}

	return nil
}

func (cfg *Config) getRunsId(runWs *workspace.RootWorkspace, planHash string) (string, error) {

	var runId string
	var createNew bool
	const hashLength = 8

	ts := cfg.Timestamp
	t := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d",
		ts.Year(), ts.Month(), ts.Day(),
		ts.Hour(), ts.Minute(), ts.Second())

	latestFilePath := runWs.GetLatestFilePath()
	reader, err := cfg.FileSystem.Open(latestFilePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// file exist but fail to open
			return "", errors.Join(err, fmt.Errorf("fail to open latest run ID"))
		}

		// if file not exist, create new run ID
		createNew = true
	} else {
		defer func() { _ = reader.Close() }()

		err = file.LoadAsText(&runId, reader)
		if err != nil {
			return "", err
		}

		// length should be >= <timestamp> + "-" + <hash 8 chars>
		if len(runId) < len(t)+1+hashLength {
			return "", fmt.Errorf("short run ID in latest file: %q", runId)
		}

		// pick up hash from runId string
		s := len(t) + 1
		hash := strings.Trim(runId[s:s+hashLength], " \t\r\n")

		// if file exist but different plan, create new run ID
		if hash != planHash[0:hashLength] {
			createNew = true
		}
	}

	if createNew {
		// Create Run ID
		runId = t + "-" + planHash[0:hashLength]
	}

	return runId, nil
}

func (cfg *Config) NewRecipeFinder(recipes []plan.RecipeFingerprint) (cmap.RecipeFinder, error) {

	finder, err := cmap.NewCapabilityMapFrom(
		system.NewSystemRecipeLoader(), cfg.FileSystem, cfg.RecipeDirs...)
	if err != nil {
		return nil, err
	}

	for _, r := range recipes {
		if e := finder.ValidateHash(r.Fqid, r.Hash, cfg.StrictCheck, cfg.FileSystem); e != nil {
			err = errors.Join(err, fmt.Errorf("recipe hash validation failed for %q: %w", r.Fqid, e))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("strict check failed: %w", err)
	}

	return finder, nil
}

func (cfg *Config) LoadState(runsWs *workspace.RunWorkspace, pln *plan.Plan) (*stateManager, error) {

	f, err := cfg.FileSystem.Open(runsWs.GetStateFilePath())
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to open state file"))
	}
	defer func() { _ = f.Close() }()

	s, err := state.LoadAsJson(f)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to load state file"))
	}

	if !canApply(s, pln) {
		return nil, fmt.Errorf("the plan is not runnable with the current state. Please re-generate the plan again with the current state again")
	}

	// if the state is same as initial state in plan, reset tasks in active apply
	if s.Hash == pln.States.Initial.Hash {
		s.Status.ActiveApply.ResetTasks()
	}
	s.Status.ActiveApply.Set(system.LaunchAt, runsWs.RunsId, pln.Hash)

	// state manager

	var opts []func(*stateManager)

	stateFile := file.NewFile(runsWs.GetStateFilePath(), cfg.FileSystem)
	opts = append(opts, WithStateSaving(stateFile))
	if prefix := cfg.DevOption.PatchPrefix(); prefix != "" {
		opts = append(opts, WithPatchPrefix(prefix, cfg.FileSystem))
	}

	stateManager := NewStateManager(
		state.NewAccessor(s),
		opts...)

	return stateManager, nil
}

func canApply(s *state.State, p *plan.Plan) bool {
	// accept case1: plan is generated from the same state
	// accept case2: status is updated with the same plan
	if s.Hash == p.States.Initial.Hash || s.Status.ActiveApply.PlanHash == p.Hash {
		return true
	}
	return false
}
