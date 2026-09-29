// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import (
	"fmt"
)

type Binder struct {
	generators map[string]*Generator
	bounds     map[string]bool
}

func NewBinder(generators map[string]*Generator) *Binder {
	return &Binder{
		generators: generators,
		bounds:     make(map[string]bool),
	}
}

func (f *Binder) Bind(selector *Selector) (string, string, error) {

	find := func(genName string, instances map[string]*Instance) (string, bool) {
		for k, i := range instances {
			id := MakeInstanceRef(genName, k)
			if _, bound := f.bounds[id]; bound || i.NodeRef != "" {
				// already bound, so search next candidate
				continue
			}

			f.bounds[id] = true
			return k, true
		}
		return "", false
	}

	generators := []string{}
	if selector.Generator != "" {
		generators = append(generators, selector.Generator)
	} else {
		for genName := range f.generators {
			generators = append(generators, genName)
		}
	}

	for _, genName := range generators {
		gen := f.generators[genName]
		if gen == nil {
			return "", "", fmt.Errorf("generator %s not found in state", genName)
		}

		instanceName, found := find(genName, gen.Instances)
		if found {
			return genName, instanceName, nil
		}
	}

	return "", "", fmt.Errorf("no available instance found in generator instances: generators=%v: selector=%v", generators, selector)
}
