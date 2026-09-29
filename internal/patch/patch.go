// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

type Patch interface {
	Apply(src Patchable) error
	GetPath() string
}

type Patchable interface {
	Add(path []string, value any) error
	Remove(path []string) error

	Update(other Patchable) error
}
