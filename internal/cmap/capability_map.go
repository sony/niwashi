// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
)

type RecipeFinder interface {
	FindRecipe(n string) (*recipe.Recipe, error)
	FindRecipeAsAdapter(n string) (*recipe.Recipe, error)
}

type CapabilityMap struct {
	// recipes[ID][Version] = Entry
	recipes map[string]map[string]*recipe.Recipe

	capabilityAlias *Alias
	toolAlias       *Alias
	fs              file.Opener
	suspects        []string
}

func NewCapabilityMap(fs file.Opener) *CapabilityMap {
	return &CapabilityMap{
		recipes:         make(map[string]map[string]*recipe.Recipe),
		capabilityAlias: NewAlias(),
		toolAlias:       NewAlias(),
		fs:              fs,
	}
}

func NewCapabilityMapFrom(loader Loader, fs file.FileSystem, dirs ...string) (*CapabilityMap, error) {
	cm := NewCapabilityMap(fs)

	if loader != nil {
		if err := cm.LoadFromLoader(loader); err != nil {
			return nil, err
		}
	}

	if len(dirs) == 0 {
		logger.Warn("No recipe directories specified for CapabilityMap")
	}

	err := cm.LoadFromDirs(dirs...)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *CapabilityMap) Add(r *recipe.Recipe) error {

	if r == nil || r.Metadata == nil {
		return fmt.Errorf("argument error: recipe is nil")
	}

	if len(r.Metadata.Id) == 0 || len(r.Metadata.Version) == 0 {
		return fmt.Errorf("format error: id=%q version=%q", r.Metadata.Id, r.Metadata.Version)
	}

	if entries, ok := cm.recipes[r.Metadata.Id]; ok {
		if foundRecipe, ok := entries[r.Metadata.Version]; ok {
			return NewConflictError(
				fmt.Sprintf("FQID=%q old=%q new=%q",
					r.Fqid(),
					foundRecipe.FilePath,
					r.FilePath))
		}
	} else {
		cm.recipes[r.Metadata.Id] = make(map[string]*recipe.Recipe)
	}

	cm.recipes[r.Metadata.Id][r.Metadata.Version] = r

	// Add Alias from spec.provides in Recipe
	var alias *Alias
	if r.Kind == recipe.KindAdapter {
		alias = cm.toolAlias
	} else {
		alias = cm.capabilityAlias
	}

	if r.Spec != nil {
		for _, pr := range r.Spec.Provides {
			alias.Add(pr.Name, r.Metadata.Id, r.Metadata.Version, pr.Attributes)
		}
	}

	return nil
}

// Get recipe from id and version constraint.
func (cm *CapabilityMap) GetRecipe(id, vc string) (*recipe.Recipe, error) {
	versions, ok := cm.recipes[id]
	if !ok {
		return nil, NewNotFoundError(fmt.Sprintf("no id found id=%q", id))
	}

	// if exact match
	if r, ok := versions[vc]; ok {
		return r, nil
	}

	// version constraints
	vs := make([]*semver.Version, len(versions))
	i := 0
	for key := range versions {
		v, err := semver.NewVersion(key)
		if err != nil {
			return nil, err
		}
		vs[i] = v
		i++
	}
	sort.Sort(sort.Reverse(semver.Collection(vs)))

	// Match version
	if vc == "latest" {
		vc = "*"
	}
	c, err := semver.NewConstraint(vc)
	if err != nil {
		return nil, err
	}
	version := ""
	for _, v := range vs {
		if c.Check(v) {
			version = v.String()
			break
		}
	}

	r, ok := versions[version]
	if !ok {
		return nil, NewNotFoundError(fmt.Sprintf("no version found version=%q", vc))
	}

	return r, nil
}

func (cm *CapabilityMap) findRecipe(n string, alias *Alias) (*recipe.Recipe, error) {

	// name pattern:
	// 1. FQID: <id>@<version|versionConstraint>
	// 2. FQID without version: <id>
	// 3. Alias: <name>
	// 4. Alias with attributes: <name>.<key>=<value>.<key2>=<value2>...

	// 1. FQID case
	id, version, err := recipe.ParseFqid(n)
	if err == nil {
		return cm.GetRecipe(id, version)
	}

	// 2. FQID without version
	id = n
	recipe, err := cm.GetRecipe(id, "*")
	if err == nil {
		return recipe, nil
	}

	// 3&4. Alias
	id, version, err = alias.Find(n)
	if err != nil {
		return nil, err
	}
	return cm.GetRecipe(id, version)
}

func (cm *CapabilityMap) FindRecipe(n string) (*recipe.Recipe, error) {

	return cm.findRecipe(n, cm.capabilityAlias)
}

func (m *CapabilityMap) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	r, err := m.findRecipe(n, m.toolAlias)
	if err != nil {
		return nil, err
	}

	if r.Kind != recipe.KindAdapter {
		return nil, fmt.Errorf("recipe is not an adapter FQID=%q", r.Fqid())
	}

	return r, nil
}

func (m *CapabilityMap) ValidateHash(fqid, hash string, strictCheck bool, fs file.FileSystem) error {
	// parse fqid
	id, version, err := recipe.ParseFqid(fqid)
	if err != nil {
		return fmt.Errorf("invalid FQID format: %q", fqid)
	}

	vmap, ok := m.recipes[id]
	if !ok {
		return fmt.Errorf("FQID not found for hash validation: FQID=%q", fqid)
	}
	r, ok := vmap[version]
	if !ok {
		return fmt.Errorf("FQID not found for hash validation: FQID=%q", fqid)
	}

	computedHash := r.Hash
	if computedHash == "" {
		// compute hash
		f, err := fs.Open(r.FilePath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		base := filepath.Dir(r.FilePath)
		rWithHash, err := recipe.LoadWithHash(
			f,
			func(path string) (io.ReadCloser, error) {
				return fs.Open(filepath.Join(base, path))
			},
			r.FilePath,
		)
		if err != nil {
			return err
		}
		computedHash = rWithHash.Hash
	}

	if computedHash != hash {
		if strictCheck {
			// return error
			return fmt.Errorf("hash mismatch for FQID=%q: expected=%q actual=%q", fqid, hash, computedHash)
		} else {
			// log warning and continue
			logger.Warn(fmt.Sprintf("hash mismatch for FQID=%q: expected=%q actual=%q", fqid, hash, computedHash))
		}
	}

	return nil
}

func (cm *CapabilityMap) addSuspect(path string) {
	for _, p := range cm.suspects {
		if p == path {
			return
		}
	}
	cm.suspects = append(cm.suspects, path)
}

func (cm *CapabilityMap) removeSuspect(path string) {
	for i, p := range cm.suspects {
		if p == path {
			cm.suspects = append(cm.suspects[:i], cm.suspects[i+1:]...)
			return
		}
	}
}

func (cm *CapabilityMap) getUnloadedSuspects() []string {
	// remove assets files from suspects
	for _, v := range cm.recipes {
		for _, r := range v {
			cm.removeSuspect(r.FilePath)
			dir := filepath.Dir(r.FilePath)
			for _, a := range r.Spec.Assets {
				cm.removeSuspect(filepath.Join(dir, a))
			}
		}
	}
	return cm.suspects
}
