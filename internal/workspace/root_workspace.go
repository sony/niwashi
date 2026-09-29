// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import "path/filepath"

const rootDirKey = "RootDir"

var (
	rootDirVar     = pvar(rootDirKey)
	storeRootPath  = plist(rootDirVar, pvalue("store"))
	stateRootPath  = plist(rootDirVar, pvalue("state"))
	runsRootPath   = plist(rootDirVar, pvalue("runs"))
	stateFilePath  = plist(stateRootPath, pvalue("state.json"))
	latestFilePath = plist(runsRootPath, pvalue("latest"))

	dryRunRootPath = plist(runsRootPath, pvalue("dry-run"))
)

type RootWorkspace struct {
	RootDir, Separator string
}

func NewRootWorkspace(rootDir string) *RootWorkspace {
	return &RootWorkspace{
		RootDir:   rootDir,
		Separator: string(filepath.Separator),
	}
}

func (r *RootWorkspace) Clone() *RootWorkspace {
	return &RootWorkspace{
		RootDir:   r.RootDir,
		Separator: r.Separator,
	}
}

func (r *RootWorkspace) Find(key string) string {
	switch key {
	case rootDirKey:
		return r.RootDir
	default:
		return ""
	}
}

func (r *RootWorkspace) ResetRootDir(rootDir string) {
	r.RootDir = rootDir
}

func (r *RootWorkspace) ResetSeparator(separator string) {
	r.Separator = separator
}

func (r *RootWorkspace) Join(paths ...string) string {
	return join(r.Separator, paths...)
}

func (r *RootWorkspace) Render(t string) string {
	return render(t, r)
}

func (r *RootWorkspace) GetRootDirPath() string {
	return rootDirVar.ToString(r)
}

func (r *RootWorkspace) GetStateRootPath() string {
	return stateRootPath.ToString(r)
}

func (r *RootWorkspace) GetStoreRootPath() string {
	return storeRootPath.ToString(r)
}

func (r *RootWorkspace) GetRunsRootPath() string {
	return runsRootPath.ToString(r)
}

func (r *RootWorkspace) GetStateFilePath() string {
	return stateFilePath.ToString(r)
}

func (r *RootWorkspace) GetDryRunRootPath() string {
	return dryRunRootPath.ToString(r)
}

func (r *RootWorkspace) GetLatestFilePath() string {
	return latestFilePath.ToString(r)
}
