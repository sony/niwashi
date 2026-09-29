// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import (
	"github.com/sony/niwashi/internal/file"
)

type Path interface {
	ToString(resolver PathResolver) string
	List() []Path
}

type pathList struct {
	paths []Path
}

func (p *pathList) ToString(resolver PathResolver) string {
	var parts []string
	for _, path := range p.paths {
		parts = append(parts, path.ToString(resolver))
	}
	return resolver.Join(parts...)
}

func (p *pathList) List() []Path {
	return p.paths
}

type pathValue struct {
	path string
}

func (p *pathValue) ToString(_ PathResolver) string {
	return p.path
}

func (p *pathValue) List() []Path {
	return []Path{p}
}

type varPath struct {
	key string
}

func (p *varPath) ToString(resolver PathResolver) string {
	return resolver.Find(p.key)
}

func (p *varPath) List() []Path {
	return []Path{p}
}

func plist(p ...Path) Path {
	return &pathList{paths: p}
}

func pvalue(p string) Path {
	return &pathValue{path: p}
}

func pvar(key string) Path {
	return &varPath{key: key}
}

func NewPath(paths ...Path) Path {
	return plist(paths...)
}

func SimplePath(p string) Path {
	return pvalue(p)
}

func SafePath(p string) Path {
	return pvalue(file.SanitizeFilename(p))
}
