// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package merge

type Mergable[T any] interface {
	Clone() T
	Merge(other T) (T, error)
}

func StringMapT[T Mergable[T]](a, b map[string]T) (map[string]T, error) {
	// make copy of a
	merged := make(map[string]T)
	for k, v := range a {
		merged[k] = v.Clone()
	}

	// Merge b into a
	for k, v := range b {
		if _, exist := a[k]; !exist {
			// Add new entry if not exist in a
			merged[k] = v.Clone()
		} else {
			var err error
			// Merge existing entry if exist in a
			merged[k], err = merged[k].Merge(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return merged, nil
}
