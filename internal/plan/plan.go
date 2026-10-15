// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import (
	"fmt"
	"slices"

	"github.com/sony/niwashi/internal/state"
)

const (
	Version = "nws.plan/v2"
)

type Mode string

const (
	ModeDefault Mode = "default"
	ModeDestroy Mode = "destroy"
	ModePrune   Mode = "prune"
)

type Plan struct {
	Version        string                 `json:"version" yaml:"version"`
	Metadata       *Metadata              `json:"metadata" yaml:"metadata"`
	Mode           Mode                   `json:"mode" yaml:"mode"`
	ExecutionPlan  ExecutionPlan          `json:"plan" yaml:"plan"`
	States         States                 `json:"states" yaml:"states"`
	Diff           map[string]*state.Diff `json:"diff" yaml:"diff"`
	Runtime        Runtime                `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	PendingUpdates []string               `json:"pendingUpdates,omitempty" yaml:"pendingUpdates,omitempty"`
	FilePath       string                 `json:"-" yaml:"-"`
	Hash           string                 `json:"-" yaml:"-"`
}

type ExecutionPlan struct {
	ConstructGraph map[string]*Workflow `json:"construct,omitempty" yaml:"construct,omitempty"`
	DestructGraph  map[string]*Workflow `json:"destruct,omitempty" yaml:"destruct,omitempty"`
}

func NewPlan() *Plan {
	p := &Plan{
		Version:  Version,
		Mode:     ModeDefault,
		Metadata: NewMetadata(),
	}
	p.SetDefaults()
	return p
}

func (p *Plan) SetDefaults() {
	p.States.SetDefaults()
	p.Runtime.SetDefaults()
}

func (p *Plan) Validate() error {

	if p.Version != Version {
		return fmt.Errorf("unsupported plan version: %s (expected: %s)", p.Version, Version)
	}

	stages, err := p.StageOrder()
	if err != nil {
		return err
	}

	// validate stage name in Diff
	for stage := range p.Diff {
		valid := false
		for _, s := range stages {
			if s == stage {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid stage name in Diff: %q", stage)
		}
	}

	// validate stage name in States.Target
	if p.States.Target != nil {
		for stage := range p.States.Target.Stages {
			if !slices.Contains(stages, stage) {
				return fmt.Errorf("invalid stage name in States.Target: %q", stage)
			}
		}
	}

	return nil
}

func (p *Plan) StageOrder() ([]string, error) {
	var wf map[string]*Workflow
	var reverse bool

	switch p.Mode {
	case ModeDefault:
		wf = p.ExecutionPlan.ConstructGraph
	case ModeDestroy, ModePrune:
		wf = p.ExecutionPlan.DestructGraph
		reverse = true
	default:
		return nil, fmt.Errorf("unsupported plan mode: %s", p.Mode)
	}

	var customStage []string
	for k := range wf {
		switch k {
		case state.InitStageName, state.MainStageName:
		default:
			customStage = append(customStage, k)
		}
	}

	if len(customStage) > 1 {
		return nil, fmt.Errorf("stage count must not exceed 1")
	}

	order := []string{}
	order = append(order, state.InitStageName)
	order = append(order, customStage...)
	order = append(order, state.MainStageName)

	if reverse {
		// swap order
		slices.Reverse(order)
	}

	return order, nil
}
