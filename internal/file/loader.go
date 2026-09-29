// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
)

func LoadAsYaml(v any, file io.Reader) error {
	ext := ".yaml"
	encDec, exist := encDec[ext]
	if !exist {
		return fmt.Errorf("unsupported file extension: %q", ext)
	}
	return encDec.decode(file, v)
}

func LoadAsJson(v any, file io.Reader) error {
	ext := ".json"
	encDec, exist := encDec[ext]
	if !exist {
		return fmt.Errorf("unsupported file extension: %q", ext)
	}
	return encDec.decode(file, v)
}

func LoadAsText(s *string, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	*s = string(data)
	return nil
}

func GenerateHash(file io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func load(encDec *EncDec, v any, file io.Reader, hash *string) error {

	if hash != nil {
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			return err
		}
		hashBytes := h.Sum(nil)
		*hash = hex.EncodeToString(hashBytes)

		if seeker, ok := file.(io.Seeker); ok {
			// Reset file pointer to the beginning after hashing
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return err
			}
		}
	}

	return encDec.decode(file, v)
}

func LoadWithDecoding(v any, path string, hash *string, fsys Opener) error {

	ext := filepath.Ext(path)
	encDec, exist := encDec[ext]
	if !exist {
		return fmt.Errorf("unsupported file extension: %q", ext)
	}

	file, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return load(&encDec, v, file, hash)
}
