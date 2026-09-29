// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import (
	"github.com/sony/niwashi/internal/state"
)

const (
	Version = "nws.plan/v1"
)

type Mode string

const (
	ModeDefault Mode = "default"
	ModeDestroy Mode = "destroy"
	ModePrune   Mode = "prune"
)

type Plan struct {
	Version        string        `json:"version" yaml:"version"`
	Metadata       *Metadata     `json:"metadata" yaml:"metadata"`
	Mode           Mode          `json:"mode" yaml:"mode"`
	ExecutionPlan  ExecutionPlan `json:"plan" yaml:"plan"`
	States         States        `json:"states" yaml:"states"`
	Diff           state.Diff    `json:"diff" yaml:"diff"`
	Runtime        Runtime       `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	PendingUpdates []string      `json:"pendingUpdates,omitempty" yaml:"pendingUpdates,omitempty"`
	FilePath       string        `json:"-" yaml:"-"`
	Hash           string        `json:"-" yaml:"-"`
}

type ExecutionPlan struct {
	ConstructGraph *Workflow `json:"construct,omitempty" yaml:"construct,omitempty"`
	DestructGraph  *Workflow `json:"destruct,omitempty" yaml:"destruct,omitempty"`
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
	// todo: format validation
	return nil
}
