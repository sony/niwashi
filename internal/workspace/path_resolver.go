// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sony/niwashi/internal/logger"
)

type PathResolver interface {
	Find(key string) string
	Join(paths ...string) string
	Render(template string) string
}

func render(t string, p any) string {
	result := t
	te, err := template.New("").Parse(t)
	if err != nil {
		return result
	}
	var builder strings.Builder
	err = te.Execute(&builder, p)
	if err != nil {
		logger.Error("failed to render template:", "err", err)
		return ""
	}

	return builder.String()
}

func join(sep string, paths ...string) string {
	if sep == "" {
		sep = "/"
	}
	return strings.Join(paths, sep)
}

var NilPathResolver = &nilPathResolver{}

type nilPathResolver struct{}

func (r *nilPathResolver) Find(key string) string {
	return ""
}

func (r *nilPathResolver) Join(paths ...string) string {
	return strings.Join(paths, string(filepath.Separator))
}

func (r *nilPathResolver) Render(t string) string {
	return t
}

// NewSeparatorResolver returns a PathResolver that joins path segments with
// the given separator, without resolving any variables. Use this to build
// paths for a remote host whose path separator differs from the local host's
// (e.g. "\\" for a Windows node), instead of NilPathResolver which always
// uses the local host's filepath.Separator.
func NewSeparatorResolver(separator string) PathResolver {
	return &separatorResolver{separator: separator}
}

type separatorResolver struct {
	separator string
}

func (r *separatorResolver) Find(key string) string {
	return ""
}

func (r *separatorResolver) Join(paths ...string) string {
	return join(r.separator, paths...)
}

func (r *separatorResolver) Render(t string) string {
	return t
}
