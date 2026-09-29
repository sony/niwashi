// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import (
	"path/filepath"

	"github.com/sony/niwashi/internal/file"
)

const runsIdKey = "RunsId"

var (
	runsIdVar         = pvar(runsIdKey)
	ephemeralRootPath = plist(runsRootPath, runsIdVar)
	planFilePath      = plist(ephemeralRootPath, pvalue("plan.json"))
)

type RunWorkspace struct {
	*RootWorkspace
	RunsId string
}

func NewRunWorkspace(rws *RootWorkspace, runsId string) *RunWorkspace {
	return &RunWorkspace{
		RootWorkspace: rws,
		RunsId:        runsId,
	}
}

func NewRunWs(rootDir string) *RunWorkspace {
	return &RunWorkspace{
		RootWorkspace: &RootWorkspace{
			RootDir:   rootDir,
			Separator: string(filepath.Separator),
		},
	}
}

func (r *RunWorkspace) Find(key string) string {
	switch key {
	case runsIdKey:
		return file.SanitizeFilename(r.RunsId)
	default:
		return r.RootWorkspace.Find(key)
	}
}

func (r *RunWorkspace) Clone() *RunWorkspace {
	return &RunWorkspace{
		RootWorkspace: r.RootWorkspace.Clone(),
		RunsId:        r.RunsId,
	}
}

func (r *RunWorkspace) GetEphemeralRootPath() string {
	return ephemeralRootPath.ToString(r)
}

func (r *RunWorkspace) GetPlanFilePath() string {
	return planFilePath.ToString(r)
}
