// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

type Mergable[T any] interface {
	Merge(other T) (T, error)
	Clone() T
}
