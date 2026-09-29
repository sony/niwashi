// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

const (
	ModePersistent = "persistent"
)

type Workspace struct {
	Mode string `json:"mode" yaml:"mode"`
}

func (w *Workspace) IsPersistent() bool {
	return w != nil && w.Mode == ModePersistent
}

func (w *Workspace) SetDefaults() {
	if w.Mode == "" {
		w.Mode = "ephemeral"
	}
}
