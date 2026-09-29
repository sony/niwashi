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
	StateAccessor *state.Accessor
	File          file.File
	PatchSaver    *PatchSaver
	dirtyState    bool
}

func NewStateManager(s *state.Accessor, opt ...func(*stateManager)) *stateManager {
	sm := &stateManager{
		StateAccessor: s,
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

func (sm *stateManager) State() *state.Accessor {
	return sm.StateAccessor
}

func (sm *stateManager) ApplyPatches(
	patches []patch.Patch,
	constraint patch.PathConstraint,
	validate func(*state.Accessor) error) error {

	if len(patches) == 0 {
		return nil
	}

	if constraint == nil {
		constraint = &patch.NilPathConstraint{}
	}

	// save patches for debug
	sm.PatchSaver.Add(patches)

	// clone state first and apply patches to it
	backupState := sm.StateAccessor.Clone()
	for _, patch := range patches {
		if err := constraint.Validate(patch.GetPath()); err != nil {
			return errors.Join(err, fmt.Errorf("patch path does not satisfy constraint: %s", patch.GetPath()))
		}
		if err := patch.Apply(backupState); err != nil {
			return errors.Join(err, fmt.Errorf("failed to apply patch: %+v", patch))
		}
	}

	newState := state.NewAccessor(backupState)
	if validate != nil {
		if err := validate(newState); err != nil {
			return errors.Join(err, fmt.Errorf("state validation failed after applying patches"))
		}
	}

	// replace state
	sm.StateAccessor = newState
	sm.dirtyState = true
	return nil
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

	err = sm.StateAccessor.WriteAsJson(f)
	if err != nil {
		return err
	}

	// clear flag
	sm.dirtyState = false
	return nil
}

func (sm *stateManager) saveTaskResult(
	operation string,
	job *workflow.Job,
	err error) error {

	aa := &sm.State().Status.ActiveApply
	if err != nil {
		aa.AddErrorTask(operation, job.Id, err)
	} else {
		// save completed task to state
		aa.AddCompleteTask(operation, job.Id)
	}

	e := sm.SaveState()
	if e != nil {
		// handle error if needed
		logger.Error("Failed to save state file", "error", e)
	}

	return e
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
