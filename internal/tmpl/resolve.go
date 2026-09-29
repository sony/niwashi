// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package tmpl

import "fmt"

type TemplateType[T any] interface {
	Merge(other T) (T, error)
	Clone() T
	FromField() string
}

func Resolve[T TemplateType[T]](tpl map[string]T, base T) (T, error) {
	var zero T
	derived := map[string]bool{}
	merged := base.Clone()
	from := base.FromField()

	for len(from) > 0 {
		if derived[from] {
			// circular reference
			return zero, fmt.Errorf("circular node template reference detected: %v", derived)
		}
		t, exist := tpl[from]
		if !exist {
			return zero, fmt.Errorf("node template %q not found", from)
		}

		var err error
		merged, err = t.Merge(merged)
		if err != nil {
			return zero, err
		}
		derived[from] = true
		from = t.FromField()
	}

	return merged, nil
}
