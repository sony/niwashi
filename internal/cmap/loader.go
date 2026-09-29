// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sony/niwashi/internal/catalog"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
)

const (
	RecipeFilename  = "nws-recipe.yaml"
	CatalogFilename = "nws-catalog.yaml"
)

type Loader interface {
	ListNames() []string
	Open(name string) (io.ReadSeekCloser, error)
}

func (cm *CapabilityMap) LoadFromLoader(loader Loader) error {
	// Load from loader and compute hash
	for _, name := range loader.ListNames() {
		reader, err := loader.Open(name)
		if err != nil {
			return err
		}
		defer func() { _ = reader.Close() }()

		path := "./system/" + name

		r, err := recipe.LoadWithHash(reader, nil, path)
		if err != nil {
			return err
		}

		if r.Spec != nil && len(r.Spec.Assets) > 0 {
			logger.Warn("Assets are not supported when loading recipe from loader. Assets will be ignored.", "recipe", r.Fqid())
		}

		if err = cm.Add(r); err != nil {
			return err
		}
		logger.Info("LoadRecipe", "path", r.FilePath)
	}
	return nil
}

func requireSkip(path string) bool {
	base := filepath.Base(path)
	return base[0] == '.'
}

func (cm *CapabilityMap) LoadFromDirs(rd ...string) error {
	cwd, _ := os.Getwd()

	for _, d := range rd {
		err := filepath.WalkDir(d, func(p string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if requireSkip(p) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.Type().IsRegular() {
				path := p
				if !filepath.IsAbs(p) {
					path = filepath.Join(cwd, p)
				}
				if e := cm.handleFile(path); e != nil {
					return e
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	unloadedYamls := cm.getUnloadedSuspects()
	for _, path := range unloadedYamls {
		logger.Warn("YAML file not loaded as a recipe: expected filename is nws-recipe.yaml or nws-catalog.yaml, or declare it under spec.assets if intentional", "path", path)
	}

	return nil
}

func (cm *CapabilityMap) handleFile(path string) error {

	switch filepath.Base(path) {
	case RecipeFilename:
		if e := cm.loadRecipeFromFile(path); e != nil {
			return e
		}
	case CatalogFilename:
		if e := cm.loadFromCatalog(path); e != nil {
			return e
		}
	default:
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			cm.addSuspect(path)
		}
	}
	return nil
}

func loadRecipe(reader io.Reader, path string) (*recipe.Recipe, error) {
	r, err := recipe.Load(reader, path)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (cm *CapabilityMap) loadRecipeFromFile(p string) error {
	reader, err := cm.fs.Open(p)
	if err != nil {
		return err
	}

	r, err := loadRecipe(reader, p)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to load recipe from file: %q", p))
	}

	if err = cm.Add(r); err != nil {
		return err
	}

	logger.Info("LoadRecipe", "path", r.FilePath)

	return nil
}

func (cm *CapabilityMap) loadFromCatalog(p string) error {
	c, e := catalog.Load(p, cm.fs)
	if e != nil {
		return e
	}

	d := filepath.Dir(p)
	for _, r := range c.Recipes {
		path := filepath.Join(d, r.Path)
		err := cm.loadRecipeFromFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}
