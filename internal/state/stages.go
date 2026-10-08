// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/profile"
)

type Stages map[string]*Stage

func (s Stages) clone() Stages {
	if s == nil {
		return nil
	}

	clone := make(Stages)
	for k, v := range s {
		clone[k] = v.clone()
	}

	return clone
}

func (s Stages) setDefaults() {
	if s == nil {
		return
	}

	for stageName, stage := range s {
		if stage == nil {
			stage = &Stage{}
			s[stageName] = stage
		}
		stage.setDefaults()
	}
}

func (s Stages) merge(other Stages) (Stages, error) {

	if s == nil {
		return other.clone(), nil
	}
	if other == nil {
		return s.clone(), nil
	}

	clone := s.clone()
	var err error
	for stageName, stage := range other {
		if _, exist := clone[stageName]; exist {
			clone[stageName], err = clone[stageName].merge(stage)
			if err != nil {
				return nil, fmt.Errorf("stage %q: %w", stageName, err)
			}
		} else {
			clone[stageName] = stage.clone()
		}
	}

	return clone, nil
}

func (s Stages) validate() error {

	var err error
	for stageName, stage := range s {
		if stage != nil {
			if e := stage.validate(); e != nil {
				err = errors.Join(err, fmt.Errorf("stage %q: %w", stageName, e))
			}
		}
	}

	return err
}

func (s Stages) applyTemplate(t *Template) error {

	if s == nil || t == nil {
		return nil
	}
	for stageName, stage := range s {
		if stage != nil {
			if err := stage.applyTemplate(t); err != nil {
				return fmt.Errorf("stage %q: %w", stageName, err)
			}
		}
	}
	return nil
}

func (s Stages) normalize(f cmap.RecipeFinder) error {
	if s == nil {
		return nil
	}

	for stageName, stage := range s {
		if stage != nil {
			if err := stage.normalize(f); err != nil {
				return fmt.Errorf("stage %q: %w", stageName, err)
			}
		}
	}

	return nil
}

func (s Stages) resolveParams(prof *profile.Profile) error {
	if s == nil || prof == nil {
		return nil
	}

	// all stages
	for stageName, stage := range s {
		if stage != nil {
			if err := stage.resolveParams(prof); err != nil {
				return fmt.Errorf("stage %q: %w", stageName, err)
			}
		}
	}

	return nil
}

func (s Stages) Find(name string) (*Stage, bool) {
	if s == nil {
		return nil, false
	}
	stage, exist := s[name]
	if !exist || stage == nil {
		return nil, false
	}
	return stage, true
}
