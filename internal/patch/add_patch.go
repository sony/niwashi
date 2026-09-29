// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

type AddPatch struct {
	Op    string
	Path  string
	Value any
}

func (p *AddPatch) GetPath() string {
	return p.Path
}

func NewAddPatch(path string, value any) *AddPatch {
	return &AddPatch{
		Op:    "set",
		Path:  path,
		Value: value,
	}
}

func (p *AddPatch) Apply(src Patchable) error {
	parts, err := ParsePath(p.Path)
	if err != nil {
		return err
	}
	return src.Add(parts, p.Value)
}
