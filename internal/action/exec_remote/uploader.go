// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/workspace"
)

type fileInfo struct {
	localPath  string
	remotePath string
	perm       *os.FileMode
}

type recipeInfo struct {
	recipe          action.RecipeSpec
	remoteRecipeDir string
}

type dirInfo struct {
	localPath  string
	remotePath string
	recursive  bool
}

type Uploader struct {
	fs       file.FileSystem
	resolver workspace.PathResolver
	files    []*fileInfo
	dirs     []*dirInfo
	recipes  []*recipeInfo
}

// NewUploader creates an Uploader. resolver.Join is used whenever the
// Uploader needs to append segments to a remote path itself (walking a
// local directory, or a recipe's spec.assets entries) rather than just
// forwarding an already-composed remote path -- e.g. remoteEnv from
// RemoteExecAction.CreateRemoteWorkspace, so nested paths use the correct
// separator for the remote target OS instead of the local host's.
func NewUploader(fs file.FileSystem, resolver workspace.PathResolver) *Uploader {
	return &Uploader{fs: fs, resolver: resolver}
}

func (u *Uploader) Add(local, remote string, perm *os.FileMode) {
	u.files = append(u.files, &fileInfo{
		localPath:  local,
		remotePath: remote,
		perm:       perm,
	})
}

func (u *Uploader) AddDir(local, remote string, recursive bool) {
	u.dirs = append(u.dirs, &dirInfo{
		localPath:  local,
		remotePath: remote,
		recursive:  recursive,
	})
}

func (u *Uploader) AddEmptyDir(remote ...string) {
	for _, r := range remote {
		u.dirs = append(u.dirs, &dirInfo{
			localPath:  "",
			remotePath: r,
			recursive:  false,
		})
	}
}

func (u *Uploader) AddRecipeAssets(r action.RecipeSpec, remoteDir string) {
	if r == nil {
		return
	}

	u.recipes = append(u.recipes, &recipeInfo{
		recipe:          r,
		remoteRecipeDir: remoteDir,
	})
}

func (u *Uploader) Upload(ctx context.Context, session transport.RemoteSession) error {

	for _, d := range u.dirs {
		if err := u.uploadDir(ctx, session, d); err != nil {
			return err
		}
	}

	for _, f := range u.files {
		if err := u.uploadFile(ctx, session, f); err != nil {
			return err
		}
	}

	for _, r := range u.recipes {
		if err := u.uploadRecipeAsset(ctx, session, r); err != nil {
			return err
		}
	}

	return nil
}

func (u *Uploader) uploadFile(ctx context.Context, session transport.RemoteSession, f *fileInfo) error {
	err := session.Upload(ctx, f.localPath, f.remotePath, f.perm)
	if err != nil {
		logger.Error("Failed to upload file", "localPath", f.localPath, "remotePath", f.remotePath, "error", err)
	}
	return err
}

func (u *Uploader) uploadDir(ctx context.Context, session transport.RemoteSession, d *dirInfo) error {

	if err := session.MkdirAll(ctx, d.remotePath); err != nil {
		return err
	}

	if !d.recursive {
		// don't walk, just create the directory
		return nil
	}

	// If local dir does not exist, nothing to upload
	if _, statErr := os.Stat(d.localPath); os.IsNotExist(statErr) {
		return nil
	}

	// walk local dir
	localPathLen := len(d.localPath)
	err := filepath.WalkDir(d.localPath, func(p string, info os.DirEntry, ie error) error {
		if ie != nil {
			return ie
		}

		remotePath := u.joinRemote(d.remotePath, p[localPathLen:])
		if info.IsDir() {
			if err := session.MkdirAll(ctx, remotePath); err != nil {
				return err
			}
		} else if info.Type().IsRegular() {
			fi, err := info.Info()
			if err != nil {
				return err
			}
			perm := fi.Mode().Perm()
			if err := session.Upload(ctx, p, remotePath, &perm); err != nil {
				return err
			}
		} else {
			logger.Warn("Skipping non-regular file", "path", p)
		}

		return nil
	})

	return err
}

func (u *Uploader) uploadRecipeAsset(ctx context.Context, session transport.RemoteSession, r *recipeInfo) error {

	assets := r.recipe.GetAssets()
	if len(assets) == 0 {
		// no assets
		return nil
	}

	localDir := r.recipe.GetDir()
	remoteDir := r.remoteRecipeDir

	for _, a := range assets {
		localPath := filepath.Join(localDir, a)

		fi, err := u.fs.Stat(localPath)
		if err != nil {
			return err
		}

		// a (e.g. "bin/setup.sh") is a recipe-authored spec.assets entry,
		// always "/"-separated regardless of host OS.
		segments := splitRelPath(a)
		remotePath := u.resolver.Join(append([]string{remoteDir}, segments...)...)
		remoteParentPath := u.resolver.Join(append([]string{remoteDir}, segments[:len(segments)-1]...)...)
		if err := session.MkdirAll(ctx, remoteParentPath); err != nil {
			return err
		}

		mode := fi.Mode()
		if err := session.Upload(ctx, localPath, remotePath, &mode); err != nil {
			return err
		}
	}
	return nil
}

// joinRemote appends localRelPath -- a path relative to some local root,
// delimited by the local host's separator (e.g. from filepath.WalkDir) --
// onto remoteBase using the remote resolver's separator, so nested
// directories translate correctly for the remote target OS.
func (u *Uploader) joinRemote(remoteBase, localRelPath string) string {
	segments := splitRelPath(localRelPath)
	if len(segments) == 0 {
		return remoteBase
	}
	return u.resolver.Join(append([]string{remoteBase}, segments...)...)
}

// splitRelPath splits a relative path into segments, accepting either the
// local host's separator (e.g. from filepath.WalkDir) or "/" (e.g. a
// recipe's spec.assets entries, which are always "/"-separated regardless
// of host OS).
func splitRelPath(p string) []string {
	var segments []string
	for _, s := range strings.Split(filepath.ToSlash(p), "/") {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}
