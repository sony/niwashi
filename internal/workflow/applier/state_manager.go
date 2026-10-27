// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
)

type stateManager struct {
	state      *state.State
	File       file.File
	PatchSaver *PatchSaver
	dirtyState bool
}

func NewStateManager(s *state.State, opt ...func(*stateManager)) *stateManager {
	sm := &stateManager{
		state: s,
	}

	for _, opt := range opt {
		opt(sm)
	}

	return sm
}

func WithPatchPrefix(prefix string, fs file.FileSystem) func(*stateManager) {
	return func(sm *stateManager) {
		if prefix == "" {
			sm.PatchSaver = nil
			return
		}

		if sm.PatchSaver == nil {
			sm.PatchSaver = NewPatchSaver(fs, prefix)
		} else {
			sm.PatchSaver.PatchPrefix = prefix
			sm.PatchSaver.fs = fs
		}
	}
}

func WithStateSaving(f file.File) func(*stateManager) {
	return func(sm *stateManager) {
		sm.File = f
	}
}

func (sm *stateManager) State() *state.State {
	return sm.state
}

func (sm *stateManager) stageView(stageName string) *state.StageView {
	sv, err := sm.state.StageView(stageName)
	if err != nil {
		return state.NewEmptyStageView(sm.state.Runtime)
	}
	return sv
}

func (sm *stateManager) StageView() *state.StageView {
	return sm.stageView(state.DefaultStageName)
}

func (sm *stateManager) applyPatches(
	stageName string,
	patches []patch.Patch,
	constraint patch.PathConstraint,
	validate func(*state.StageView) error) error {

	if len(patches) == 0 {
		return nil
	}

	if constraint == nil {
		constraint = &patch.NilPathConstraint{}
	}

	// save patches for debug
	sm.PatchSaver.Add(patches)

	// clone state first and apply patches to it
	backupState := sm.state.Clone()

	// create stage view for applying patches
	stage, err := backupState.StageForWrite(stageName)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to get stage for write"))
	}
	sv := backupState.MakeStageView(stage)

	// apply all patches
	for _, patch := range patches {
		if err := constraint.Validate(patch.GetPath()); err != nil {
			return errors.Join(err, fmt.Errorf("patch path does not satisfy constraint: %s", patch.GetPath()))
		}
		if err := patch.Apply(sv); err != nil {
			return errors.Join(err, fmt.Errorf("failed to apply patch: %+v", patch))
		}
	}

	// update the backup state with the modified stage view
	if stageName != state.InitStageName {
		stage.Inventory = sv.Inventory
		stage.Infrastructure = sv.Infrastructure
	}
	backupState.Runtime = sv.Runtime

	if validate != nil {
		if err := validate(sv); err != nil {
			return errors.Join(err, fmt.Errorf("state validation failed after applying patches"))
		}
	}

	// replace state
	sm.state = backupState
	sm.dirtyState = true
	return nil
}

func (sm *stateManager) ApplyPatches(
	patches []patch.Patch,
	constraint patch.PathConstraint,
	validate func(*state.StageView) error) error {
	return sm.applyPatches(
		state.DefaultStageName,
		patches,
		constraint,
		validate,
	)
}

func (sm *stateManager) IsStateDirty() bool {
	return sm.dirtyState
}

func (sm *stateManager) SaveState() (err error) {

	if sm.File == nil {
		// do nothing
		logger.Debug("state file is nil, skipping state save")
		return nil
	}

	if !sm.dirtyState {
		// don't save if state is not dirty
		return nil
	}

	f, err := sm.File.OpenFile(os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()

	// compact the state by removing empty stages
	sm.state.Compact()

	err = sm.state.WriteAsJson(f)
	if err != nil {
		return err
	}

	// clear flag
	sm.dirtyState = false
	return nil
}

func (sm *stateManager) saveTaskResult(
	operation string,
	stageName string,
	job *workflow.Job,
	err error) error {

	id := stageName + "/" + job.Id
	aa := &sm.State().Status.ActiveApply
	if err != nil {
		aa.AddErrorTask(operation, id, err)
	} else {
		// save completed task to state
		aa.AddCompleteTask(operation, id)
	}

	e := sm.SaveState()
	if e != nil {
		// handle error if needed
		logger.Error("Failed to save state file", "error", e)
	}

	return e
}

func (sm *stateManager) setExports(stageName string, exports state.Exports) error {
	stage, err := sm.state.StageForWrite(stageName)
	if err != nil {
		return err
	}
	stage.Export = exports
	sm.dirtyState = true
	return nil
}

func (sm *stateManager) SaveTaskResult(
	operation string,
	job *workflow.Job,
	err error) error {
	return sm.saveTaskResult(
		operation,
		state.DefaultStageName,
		job,
		err,
	)
}

type stageStateManager struct {
	*stateManager
	stageName string
}

func (sm *stageStateManager) StageView() *state.StageView {
	return sm.stageView(sm.stageName)
}

func (sm *stageStateManager) ApplyPatches(
	patches []patch.Patch,
	constraint patch.PathConstraint,
	validate func(*state.StageView) error) error {
	return sm.applyPatches(
		sm.stageName,
		patches,
		constraint,
		validate,
	)
}

func (sm *stageStateManager) SaveTaskResult(
	operation string,
	job *workflow.Job,
	err error) error {
	return sm.saveTaskResult(
		operation,
		sm.stageName,
		job,
		err,
	)
}

func (sm *stageStateManager) StageName() string {
	return sm.stageName
}

func (sm *stageStateManager) SetExports(exports state.Exports) error {
	return sm.setExports(sm.stageName, exports)
}

type PatchSaver struct {
	fs           file.FileSystem
	PatchPrefix  string
	PatchCounter int
}

func NewPatchSaver(fs file.FileSystem, prefix string) *PatchSaver {
	return &PatchSaver{
		fs:          fs,
		PatchPrefix: prefix,
	}
}

func (ps *PatchSaver) Add(patches []patch.Patch) {
	if ps == nil || ps.PatchPrefix == "" {
		return // do nothing
	}

	jsonData, _ := json.Marshal(patches)
	filepath := fmt.Sprintf("%s%03d.json", ps.PatchPrefix, ps.PatchCounter)
	err := file.WriteFile(ps.fs, filepath, jsonData, 0644)
	if err != nil {
		logger.Warn("failed to save patch file", "error", err)
	}
	ps.PatchCounter++
}
