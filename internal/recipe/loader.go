// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/sony/niwashi/internal/file"
)

func Load(reader io.Reader, path string) (*Recipe, error) {
	var r Recipe
	err := file.LoadAsYaml(&r, reader)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to load recipe"))
	}
	r.FilePath = path

	r.SetDefaults()
	if err = r.Validate(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("validation error recipe %q", r.Fqid()))
	}

	return &r, nil
}

func LoadWithHash(
	reader io.Reader,
	assetOpener func(path string) (io.ReadCloser, error),
	path string,
) (*Recipe, error) {
	h := sha256.New()

	tee := io.TeeReader(reader, h)
	var r Recipe
	if err := file.LoadAsYaml(&r, tee); err != nil {
		return nil, errors.Join(err, fmt.Errorf("fail to load recipe"))
	}
	r.FilePath = path

	if assetOpener != nil {
		// compute hash of assets
		for _, assetPath := range r.Spec.Assets {
			f, err := assetOpener(assetPath)
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("fail to open asset: %q", assetPath))
			}

			_, err = io.Copy(h, f)
			err = errors.Join(err, f.Close())
			if err != nil {
				return nil, errors.Join(err, fmt.Errorf("fail to read asset: %q", assetPath))
			}
		}
	}

	r.SetDefaults()
	if err := r.Validate(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("validation error recipe %q", r.Fqid()))
	}

	r.Hash = hex.EncodeToString(h.Sum(nil))

	return &r, nil
}
