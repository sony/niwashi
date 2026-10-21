// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"fmt"
	"maps"
	"slices"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/types"
)

type State struct {
	Version  string     `json:"version" yaml:"version"`
	Metadata types.Dict `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Template *Template  `json:"template,omitempty" yaml:"template,omitempty"`
	Stage    `yaml:",inline"`
	Runtime  *Runtime `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Status   *Status  `json:"status,omitempty" yaml:"status,omitempty"`
	Hash     string   `json:"-" yaml:"-"`
	Stages   Stages   `json:"stages,omitempty" yaml:"stages,omitempty"`
}

func NewState() *State {
	state := &State{
		Version: "nws.state/v1",
	}

	state.SetDefaults()

	return state
}

func (s *State) Clone() *State {
	if s == nil {
		return nil
	}

	clone := &State{}
	clone.Version = s.Version
	clone.Metadata = s.Metadata.Clone()
	clone.Template = s.Template.Clone()
	clone.Runtime = s.Runtime.Clone()
	clone.Status = s.Status.Clone()
	clone.Stage = *s.Stage.clone() //nolint:staticcheck
	clone.Stages = s.Stages.clone()

	return clone
}

func (s *State) SetDefaults() {

	if s.Template == nil {
		s.Template = s.Template.SetDefaults()
	}

	// fill minimal required fields
	s.Stage.setDefaults() //nolint:staticcheck

	if s.Stages == nil {
		s.Stages = Stages{}
	}
	s.Stages.setDefaults()

	if s.Runtime == nil {
		s.Runtime = &Runtime{}
	}
	s.Runtime.SetDefaults()

	if s.Status == nil {
		s.Status = &Status{}
	}
	s.Status.SetDefaults()
}

func (s *State) Merge(other *State) (*State, error) {

	if s == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return s.Clone(), nil
	}

	// Merge only the following fields:
	// - Template
	// - Stage (the main stage)
	// - Stages
	clone := &State{
		Version:  s.Version,
		Metadata: s.Metadata.Clone(),
		Runtime:  s.Runtime.Clone(),
		Status:   s.Status.Clone(),
	}

	var err error
	clone.Template, err = s.Template.Merge(other.Template)
	if err != nil {
		return nil, err
	}

	// "main" stage
	stage, err := s.Stage.merge(&other.Stage) //nolint:staticcheck
	if err != nil {
		return nil, err
	}
	clone.Stage = *stage

	// Merge other stages
	clone.Stages, err = s.Stages.merge(other.Stages)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

func (s *State) ApplyTemplates() error {
	if s == nil {
		return nil
	}

	if err := s.Stage.applyTemplate(s.Template); err != nil { //nolint:staticcheck
		return err
	}

	if err := s.Stages.applyTemplate(s.Template); err != nil {
		return err
	}

	return nil
}

func (s *State) Normalize(f cmap.RecipeFinder) error {
	if s == nil {
		return nil
	}

	if f == nil {
		return fmt.Errorf("RecipeFinder is required for normalization")
	}

	// "main" stage
	if err := s.Stage.normalize(f); err != nil { //nolint:staticcheck
		return err
	}

	// other stages
	if s.Stages != nil {
		if err := s.Stages.normalize(f); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) ResolveParams(prof *profile.Profile) error {
	if s == nil {
		return nil
	}
	if prof == nil {
		return nil
	}

	// "main" stage
	if err := s.Stage.resolveParams(prof); err != nil { //nolint:staticcheck
		return err
	}

	// other stages
	if s.Stages != nil {
		if err := s.Stages.resolveParams(prof); err != nil {
			return err
		}
	}

	return nil
}

func effectiveExports(target, initial *Stage) Exports {
	var exports Exports
	if initial != nil {
		exports = initial.Export.Clone()
	}
	if target != nil {
		if exports == nil {
			exports = target.Export.Clone()
		} else {
			for k, v := range target.Export {
				exports[k] = v.Clone()
			}
		}
	}
	return exports
}

func (s *State) ValidateExportReferences(initial *State) error {
	if s == nil {
		return nil
	}

	// collect all stage names
	stageNames := map[string]struct{}{}
	for name := range s.Stages {
		stageNames[name] = struct{}{}
	}
	if initial != nil {
		for name := range initial.Stages {
			stageNames[name] = struct{}{}
		}
	}

	exports := map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(stageNames)) {
		targetStage, _ := s.FindStage(name)
		initialStage, _ := initial.FindStage(name)

		effective := effectiveExports(targetStage, initialStage)
		if err := effective.ValidateWith(targetStage, initialStage); err != nil {
			return fmt.Errorf("stage %q: %w", name, err)
		}

		for _, exportName := range slices.Sorted(maps.Keys(effective)) {
			if other, exists := exports[exportName]; exists {
				return fmt.Errorf("export %q is defined in both stage %q and stage %q", exportName, other, name)
			}
			exports[exportName] = name
		}
	}

	return nil
}

func (s *State) ValidatePrunedExportReferences(initial *State) error {
	if s == nil || initial == nil {
		return nil
	}

	for _, name := range slices.Sorted(maps.Keys(initial.Stages)) {
		initialStage := initial.Stages[name]
		if initialStage == nil || initialStage.Export == nil {
			continue
		}

		targetStage, _ := s.FindStage(name)

		// the whole stage is removed by prune, so its exports need no check
		if targetStage == nil {
			continue
		}

		for _, exportName := range slices.Sorted(maps.Keys(initialStage.Export)) {
			if _, exists := targetStage.Export[exportName]; !exists {
				// the export is removed by prune, so it needs no check
				continue
			}

			// the export remains after prune, so its source must remain too
			if err := initialStage.Export[exportName].ValidateWith(targetStage); err != nil {
				return fmt.Errorf("stage %q: export %q: removed by prune: %w", name, exportName, err)
			}
		}

	}

	return nil
}

func validateStageName(name string) error {
	if !types.StageNamePattern.MatchString(name) {
		return fmt.Errorf("invalid stage name: %s", name)
	}
	return nil
}

func (s *State) StageForWrite(name string) (*Stage, error) {
	switch name {
	case MainStageName:
		return &s.Stage, nil
	case InitStageName:
		return &Stage{}, nil
	default:
		if err := validateStageName(name); err != nil {
			return nil, err
		}
		if s.Stages == nil {
			s.Stages = make(map[string]*Stage)
		}

		if stage, exists := s.Stages[name]; exists {
			return stage, nil
		}
		stage := &Stage{}
		stage.setDefaults()
		s.Stages[name] = stage
		return stage, nil
	}
}

func (s *State) FindStage(name string) (*Stage, bool) {
	if s == nil {
		return nil, false
	}

	switch name {
	case MainStageName:
		return &s.Stage, true
	case InitStageName:
		return &Stage{}, true
	default:
		if s.Stages != nil {
			return s.Stages.Find(name)
		}
	}

	return nil, false
}

func (s *State) StageView(name string) (*StageView, error) {
	if s == nil {
		return nil, fmt.Errorf("state is nil")
	}

	stage, exist := s.FindStage(name)
	if !exist {
		return nil, fmt.Errorf("unsupported stage name: %s", name)
	}
	return s.makeStageView(stage), nil
}

func (s *State) MakeStageView(stage *Stage) *StageView {
	return s.makeStageView(stage)
}

func (s *State) makeStageView(stage *Stage) *StageView {
	return &StageView{
		Inventory:      stage.Inventory,
		Infrastructure: stage.Infrastructure,
		Runtime:        s.Runtime,
	}
}

func (s *State) Compact() {
	if s == nil {
		return
	}

	// Remove empty stages from the stages
	for name, stage := range s.Stages {
		if stage.isEmpty() {
			delete(s.Stages, name)
		}
	}
}
