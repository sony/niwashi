// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

import (
	"fmt"
	"strconv"

	"github.com/sony/niwashi/internal/logger"
)

func RemoveSimpleValue(path []string, remover func()) error {
	if len(path) != 0 {
		return fmt.Errorf("access does not support sub-paths")
	}

	remover()
	return nil
}

func RemoveListValue(
	path []string,
	remover func(index int) error,
	clearAll func(),
) error {

	if len(path) == 0 {
		clearAll()
	} else if len(path) == 1 {
		var index int
		if path[0] == "-" {
			index = -1
		} else {
			var err error
			index, err = strconv.Atoi(path[0])
			if err != nil {
				return fmt.Errorf("unsupported list removal path: %s", path[0])
			}
		}
		return remover(index)
	} else {
		return fmt.Errorf("invalid path for list removal: %v", path)
	}

	return nil
}
func RemoveCompositeMap[V Patchable](
	path []string,
	mapStore map[string]V,
) error {

	if len(path) == 0 {
		return fmt.Errorf("no key specified in path")
	}

	if len(path) == 1 {
		delete(mapStore, path[0])
		return nil
	}

	key := path[0]
	_, found := mapStore[key]
	if !found {
		logger.Warn("key not found in map:", "key", key)
		return nil
	}

	return RemoveComposite(
		path[1:],
		func() (V, bool) {
			v, found := mapStore[key]
			return v, found
		},
		func() {
			delete(mapStore, key)
		},
	)
}

func RemoveComposite[T Patchable](
	path []string,
	getter func() (T, bool),
	remover func(),
) error {

	if len(path) > 0 {
		existing, ok := getter()
		if !ok {
			return fmt.Errorf("no existing value to access sub-paths")
		}
		return existing.Remove(path)
	}

	remover()
	return nil
}
