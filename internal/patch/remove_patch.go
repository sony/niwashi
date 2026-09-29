// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

type RemovePatch struct {
	Op   string
	Path string
}

func (p *RemovePatch) GetPath() string {
	return p.Path
}

func NewRemovePatch(path string) *RemovePatch {
	return &RemovePatch{
		Op:   "remove",
		Path: path,
	}
}

func (p *RemovePatch) Apply(src Patchable) error {
	parts, err := ParsePath(p.Path)
	if err != nil {
		return err
	}

	return src.Remove(parts)
}
