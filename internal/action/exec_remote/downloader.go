// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"context"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
)

type Downloader struct {
	fs   file.FileSystem
	dirs []*dirInfo
}

func NewDownloader(fs file.FileSystem) *Downloader {
	return &Downloader{fs: fs}
}

func (d *Downloader) AddDir(local, remote string, recursive bool) {
	d.dirs = append(d.dirs, &dirInfo{
		localPath:  local,
		remotePath: remote,
		recursive:  recursive,
	})
}

func (d *Downloader) Download(ctx context.Context, session transport.RemoteSession) error {
	for _, dir := range d.dirs {
		if err := session.DownloadDir(ctx, dir.remotePath, dir.localPath, dir.recursive); err != nil {
			return err
		}
	}
	return nil
}
