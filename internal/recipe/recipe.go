// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const (
	RecipeSpecVersion = "nws.recipe/v1"

	KindHost    = "host"
	KindInfra   = "infra"
	KindNode    = "node"
	KindCluster = "cluster"
	KindAdapter = "adapter"
)

var kinds = []string{
	KindHost,
	KindInfra,
	KindNode,
	KindCluster,
	KindAdapter,
}

type Recipe struct {
	Version  string    `json:"version" yaml:"version"`
	Kind     string    `json:"kind" yaml:"kind"`
	Metadata *Metadata `json:"metadata" yaml:"metadata"`
	Spec     *Spec     `json:"spec" yaml:"spec"`
	FilePath string    `json:"-" yaml:"-"`
	Hash     string    `json:"-" yaml:"-"`
}

func (r *Recipe) Fqid() string {
	if r.Metadata == nil {
		return ""
	}
	return MakeFqid(r.Metadata.Id, r.Metadata.Version)
}

func (r *Recipe) SetDefaults() {
	if r == nil {
		return
	}
	if r.Version == "" {
		r.Version = RecipeSpecVersion
	}
	if r.Metadata == nil {
		r.Metadata = &Metadata{}
	}
	if r.Spec == nil {
		r.Spec = &Spec{}
	}

	if r.Kind == KindAdapter {
		(*Adapter)(r.Spec).SetDefaults()
	} else {
		r.Spec.SetDefaults()
	}
}

func (r *Recipe) Validate() error {
	var err error

	if r.Version != RecipeSpecVersion {
		err = errors.Join(err, fmt.Errorf("unsupported recipe version: %q", r.Version))
	}

	if !slices.Contains(kinds, r.Kind) {
		err = errors.Join(fmt.Errorf("kind=%q kind must be one of %q", r.Kind, strings.Join(kinds, ", ")))
		// don't validate any more if kind is invalid
		return err
	}

	if r.Metadata == nil {
		err = errors.Join(err, fmt.Errorf("metadata is required"))
	} else {
		if e := r.Metadata.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if r.Spec == nil {
		err = errors.Join(err, fmt.Errorf("spec is required"))
	} else {
		if r.Kind == KindAdapter {
			if e := (*Adapter)(r.Spec).Validate(); e != nil {
				err = errors.Join(err, e)
			}
		} else {
			if e := r.Spec.Validate(); e != nil {
				err = errors.Join(err, e)
			}
		}
	}

	return err
}

func (r *Recipe) GetAdapter() *Adapter {
	if r == nil || r.Spec == nil || r.Kind != KindAdapter {
		return nil
	}
	return (*Adapter)(r.Spec)
}

func MakeFqid(id, version string) string {
	return id + "@" + version
}

var fqidSubmatchPattern = regexp.MustCompile("^(" + recipeIdPatternString + ")@(.+)$")

func ParseFqid(fqid string) (string, string, error) {
	match := fqidSubmatchPattern.FindStringSubmatch(fqid)
	if len(match) == 0 {
		return "", "", fmt.Errorf("format error: invalid FQID %q", fqid)
	}
	return match[1], match[2], nil
}

func (r *Recipe) FilterTasks(f func(*Task) bool) []*Task {
	var tasks []*Task
	for _, t := range r.Spec.Tasks {
		if f(t) {
			tasks = append(tasks, t)
		}
	}
	return tasks
}
