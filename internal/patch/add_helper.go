// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

import (
	"cmp"
	"encoding/json"
	"fmt"

	"github.com/sony/niwashi/internal/logger"
)

type SimpleValue interface {
	cmp.Ordered | ~bool
}

func AddSimpleValue[T SimpleValue](path []string, value any, add func(v T)) error {
	if len(path) != 0 {
		return fmt.Errorf("access does not support sub-paths")
	}

	v, ok := value.(T)
	if !ok {
		return fmt.Errorf("value has incorrect type: expected %T", *new(T))
	}
	add(v)
	return nil
}

func ConvertValue[T any](value any) (T, error) {
	if v, ok := value.(T); ok {
		return v, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to marshal value: %w", err)
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		var zero T
		logger.Error("Cannot convert value", "data", string(data), "error", err)
		return zero, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return result, nil
}

func ConvertValuePtr[T any](value any) (*T, error) {
	if value == nil {
		return nil, nil
	}
	if v, ok := value.(*T); ok {
		return v, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error("Cannot convert value", "data", string(data), "error", err)
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return &result, nil
}

func AddValue[T Patchable](
	path []string,
	value any,
	getter func() (T, bool),
	setter func(v T),
) error {
	if len(path) == 0 {
		result, err := ConvertValue[T](value)
		if err != nil {
			return err
		}

		setter(result)
		return nil
	} else {
		existing, ok := getter()
		if !ok {
			return fmt.Errorf("no existing value to access sub-paths")
		}
		return existing.Add(path, value)
	}
}

func AddListValue[T any](
	path []string,
	value any,
	setter func(v []T),
	appender func(v []T),
) error {
	var updater func(v []T)

	if len(path) == 0 {
		updater = setter
	} else if len(path) == 1 && path[0] == "-" {
		updater = appender
	} else {
		return fmt.Errorf("only full replacement or appending is supported for lists")
	}

	v, err := ConvertValue[[]T](value)
	if err != nil {
		return err
	}

	updater(v)

	return nil
}

func AddCompositeValue[T Patchable](
	path []string,
	value any,
	getter func() (T, bool),
	setter func(v T),
) error {

	if len(path) > 0 {
		existing, ok := getter()
		if !ok {
			return fmt.Errorf("no existing value to access sub-paths")
		}
		return existing.Add(path, value)
	}

	result, err := ConvertValue[T](value)
	if err != nil {
		return err
	}

	setter(result)
	return nil
}

func AddCompositeToMap[V Patchable](
	path []string,
	value any,
	mapStore map[string]V,
) error {

	if len(path) == 0 {
		// todo: replace the whole map?
		return fmt.Errorf("no key specified in path")
	}

	if len(path) == 1 {
		key := path[0]
		result, err := ConvertValue[V](value)
		if err != nil {
			return err
		}
		if _, exists := mapStore[key]; exists {
			if err := mapStore[key].Update(result); err != nil {
				return err
			}
		} else {
			mapStore[key] = result
		}
		return nil
	}

	key := path[0]
	_, found := mapStore[key]
	if !found {
		return fmt.Errorf("key %s not found in map", key)
	}

	return AddCompositeValue(
		path[1:],
		value,
		func() (V, bool) {
			v, found := mapStore[key]
			return v, found
		},
		func(v V) {
			mapStore[key] = v
		},
	)
}
