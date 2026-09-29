// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"

	"dario.cat/mergo"
)

type Dict map[string]any

func (d *Dict) Clone() Dict {
	if d == nil || *d == nil {
		return nil
	}
	if len(*d) == 0 {
		return make(Dict)
	}
	return CloneMap(*d)
}

func (d *Dict) validateRecursive(re *regexp.Regexp, fnValidate func(val Dict) error) error {
	if d == nil || *d == nil {
		return nil
	}

	var err error
	for k, v := range *d {
		if !re.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid key: %s", k))
		}

		if e := validateKeys(v, fnValidate); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid value for key %s: %w", k, e))
		}
	}

	return err
}

func validateKeys(v any, fnValidate func(val Dict) error) error {
	switch val := v.(type) {
	case Dict:
		return fnValidate(val)
	case map[string]any:
		dict := Dict(val)
		return fnValidate(dict)
	case []any:
		var err error
		for _, item := range val {
			if e := validateKeys(item, fnValidate); e != nil {
				err = errors.Join(err, e)
			}
		}
		return err
	default:
		return nil
	}
}

func (d *Dict) ValidateAsLabelKey() error {
	return d.validateRecursive(
		LabelKeyPattern,
		func(val Dict) error {
			return val.ValidateAsLabelKey()
		})
}

func (d *Dict) ValidateAsTemplateVar() error {
	return d.validateRecursive(
		TemplateVarNamePattern,
		func(val Dict) error {
			return val.ValidateAsTemplateVar()
		})
}

func (d *Dict) ValidateAsParams() error {
	return d.validateRecursive(
		ParamsNamePattern,
		func(val Dict) error {
			return val.ValidateAsParams()
		})
}

func CloneMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = cloneValue(v)
	}
	return result
}

func cloneValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return CloneMap(val)
	case Dict:
		return Dict(CloneMap(val))
	case []any:
		return cloneSlice(val)
	case []string:
		return append([]string(nil), val...)
	default:
		return v
	}
}

func cloneSlice(s []any) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = cloneValue(v)
	}
	return result
}

func (d *Dict) Get(keys ...string) (any, bool) {
	if d == nil || *d == nil {
		return nil, false
	}

	current := Dict(*d)

	for i, k := range keys {
		value, ok := current[k]
		if !ok {
			return nil, false
		}
		if i == len(keys)-1 {
			return value, true
		}

		switch v := value.(type) {
		case Dict:
			current = v
		case map[string]any:
			current = Dict(v)
		default:
			return nil, false
		}
	}
	return nil, false
}

func (d *Dict) Set(value any, keys ...string) error {
	return d.setInternal(value, false, keys...)
}

func (d *Dict) SetPath(value any, keys ...string) error {
	return d.setInternal(value, true, keys...)
}

func (d *Dict) setInternal(value any, setpath bool, keys ...string) error {
	if d == nil || *d == nil {
		return fmt.Errorf("cannot set value on nil Dict")
	}

	current := Dict(*d)

	for i, k := range keys {
		// if last key, set value
		if i == len(keys)-1 {
			current[k] = value
			return nil
		}

		value, ok := current[k]
		if !ok {
			if setpath {
				newDict := make(Dict)
				current[k] = newDict
				current = newDict
				continue
			}
			return fmt.Errorf("key not found: %s", k)
		}

		switch v := value.(type) {
		case Dict:
			current = v
		case map[string]any:
			current = Dict(v)
		default:
			return fmt.Errorf("invalid type at key: %s", k)
		}
	}
	return fmt.Errorf("failed to set value")
}

func Merge(a, b Dict) (Dict, error) {
	c := maps.Clone(a)
	if err := mergo.Merge(&c, b, mergo.WithOverride); err != nil {
		return nil, err
	}
	return c, nil
}

func MergeMap(a, b map[string]any) (map[string]any, error) {
	xa := Dict(a)
	xb := Dict(b)

	xc, err := Merge(xa, xb)
	if err != nil {
		return nil, err
	}
	return xc, nil
}

func (d *Dict) Hash() string {
	if d == nil || *d == nil {
		return ""
	}
	// Use a consistent serialization method for hashing
	b, err := json.Marshal(*d)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}
