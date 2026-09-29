// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

import (
	"github.com/sony/niwashi/internal/system"
)

type Metadata struct {
	CreatedAt     string              `json:"createdAt" yaml:"createdAt"`
	NwsctlVersion string              `json:"nwsctlVersion" yaml:"nwsctlVersion"`
	Recipes       []RecipeFingerprint `json:"recipes" yaml:"recipes"`
}

type RecipeFingerprint struct {
	Fqid string `json:"fqid" yaml:"fqid"`
	Hash string `json:"hash" yaml:"hash"`
}

func NewMetadata() *Metadata {
	return &Metadata{
		CreatedAt:     system.LaunchAt.String(),
		NwsctlVersion: system.CliVersion.Version,
		Recipes:       make([]RecipeFingerprint, 0),
	}
}
