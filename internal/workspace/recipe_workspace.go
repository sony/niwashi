// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import "github.com/sony/niwashi/internal/file"

const scopeTargetKey = "ScopeTarget"
const recipeFqidKey = "RecipeFqid"

var (
	recipeFqidVar        = pvar(recipeFqidKey)
	scopeTargetVar       = pvar(scopeTargetKey)
	ephemeralWorkDirPath = plist(ephemeralRootPath, scopeTargetVar, recipeFqidVar)
	storeWorkDirPath     = plist(storeRootPath, scopeTargetVar, recipeFqidVar)
)

type RecipeWorkspace struct {
	*RunWorkspace
	Phase, Target, RecipeFqid string
	persistent                bool
}

func NewRecipeWorkspace(runWs *RunWorkspace, phase, target, recipeFqid string, persistent bool) *RecipeWorkspace {
	return &RecipeWorkspace{
		RunWorkspace: runWs,
		Phase:        phase,
		Target:       target,
		RecipeFqid:   recipeFqid,
		persistent:   persistent,
	}
}

func (r *RecipeWorkspace) Find(key string) string {
	switch key {
	case scopeTargetKey:
		var scopeTarget string
		if r.Target == "" {
			scopeTarget = r.Phase
		} else {
			scopeTarget = r.Phase + "-" + r.Target
		}
		return file.SanitizeFilename(scopeTarget)
	case recipeFqidKey:
		return file.SanitizeFilename(r.RecipeFqid)
	default:
		return r.RunWorkspace.Find(key)
	}
}

func (r *RecipeWorkspace) Join(paths ...string) string {
	return r.RunWorkspace.Join(paths...)
}

func (r *RecipeWorkspace) Render(t string) string {
	return render(t, r)
}

func (r *RecipeWorkspace) Clone() *RecipeWorkspace {
	return &RecipeWorkspace{
		RunWorkspace: r.RunWorkspace.Clone(),
		Phase:        r.Phase,
		Target:       r.Target,
		RecipeFqid:   r.RecipeFqid,
		persistent:   r.persistent,
	}
}

func (r *RecipeWorkspace) GetWorkDirPath() string {
	if r.persistent {
		return r.GetStoreWorkDirPath()
	} else {
		return r.GetEphemeralWorkDirPath()
	}
}

func (r *RecipeWorkspace) GetStoreWorkDirPath() string {
	return storeWorkDirPath.ToString(r)
}

func (r *RecipeWorkspace) GetEphemeralWorkDirPath() string {
	return ephemeralWorkDirPath.ToString(r)
}
