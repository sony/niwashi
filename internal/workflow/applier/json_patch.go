// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
)

type JsonPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

func (jp *JsonPatch) GetPath() string {
	return jp.Path
}

type PathConstraint interface {
	GetDefaultPrefix() string
}

func trimForPath(path string) string {
	nows := strings.Trim(path, " \t\r\n")
	return strings.TrimPrefix(nows, "/")
}

func MakeJsonPatch(
	sc *recipe.StateChangeOperation,
	tplParams any,
	fs file.FileSystem,
	constraint PathConstraint) ([]patch.Patch, error) {

	patches := []patch.Patch{}

	var absPath string
	up := trimForPath(sc.Path)
	if up == "" {
		absPath = constraint.GetDefaultPrefix()
	} else {
		absPath = constraint.GetDefaultPrefix() + "/" + trimForPath(sc.Path)
	}

	if sc.Count != "" {
		params := tplParams.(*action.TemplateParams)
		params.Loop = nil
		countStr, err := RenderTemplate(sc.Count, params)
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, err
		}
		if count < 0 || count > 10000 {
			return nil, fmt.Errorf("invalid count value: %d", count)
		}

		// loop operation
		index := 0
		params.Loop = make(map[string]int)
		for index < count {
			params.Loop["index"] = index
			value, err := makeValue(sc, params, fs)
			if err != nil {
				return nil, err
			}
			path, err := RenderTemplate(absPath, params)
			if err != nil {
				return nil, err
			}
			patches = append(patches, &JsonPatch{
				Op:    sc.Op,
				Path:  path,
				Value: value,
			})
			index++
		}
	} else {

		value, err := makeValue(sc, tplParams, fs)
		if err != nil {
			return nil, err
		}
		path, err := RenderTemplate(absPath, tplParams)
		if err != nil {
			return nil, err
		}
		p := &JsonPatch{
			Op:   sc.Op,
			Path: path,
		}
		if value != nil {
			p.Value = value
		}
		patches = append(patches, p)
	}

	return patches, nil
}

func makeValue(sc *recipe.StateChangeOperation, tplParams any, fs file.FileSystem) (any, error) {
	if sc.Value != nil {
		return sc.Value, nil
	} else if sc.ValueFromFile != "" {
		filePath, err := RenderTemplate(sc.ValueFromFile, tplParams)
		if err != nil {
			return nil, err
		}

		reader, err := fs.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer func() { _ = reader.Close() }()

		ext := filepath.Ext(filePath)

		var value any
		switch ext {
		case ".json", ".jsonc":
			err = file.LoadAsJson(&value, reader)
		case ".yaml", ".yml":
			err = file.LoadAsYaml(&value, reader)
		default:
			var v string
			err = file.LoadAsText(&v, reader)
			value = v
		}

		if err != nil {
			return nil, err
		}
		return value, nil
	} else if sc.ValueFromJson != "" {
		jsonStr, err := RenderTemplate(sc.ValueFromJson, tplParams)
		if err != nil {
			return nil, err
		}
		// decode jsonStr to value
		var value any
		err = json.Unmarshal([]byte(jsonStr), &value)
		if err != nil {
			return nil, err
		}
		return value, nil
	}

	return nil, nil
}

func (jp *JsonPatch) Apply(src patch.Patchable) error {

	path, err := patch.ParsePath(jp.Path)
	if err != nil {
		return err
	}

	switch jp.Op {
	case recipe.OpSet, recipe.OpAdd:
		return src.Add(path, jp.Value)
	case recipe.OpRemove:
		return src.Remove(path)
	default:
		return fmt.Errorf("unsupported JSON patch operation: %s", jp.Op)
	}
}
