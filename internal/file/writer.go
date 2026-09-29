// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WriteAsYaml(p any, w io.Writer) error {
	return encDec[".yaml"].encode(w, p)
}

func WriteAsJson(p any, w io.Writer) error {
	return encDec[".json"].encode(w, p)
}

func Write(data []byte, writer io.Writer) error {
	_, err := writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func WriteWithEncoding(p any, path string, fs Creator) (err error) {
	ext := filepath.Ext(path)
	encDec, exist := encDec[ext]
	if !exist {
		return fmt.Errorf("unsupported file extension: %q", ext)
	}

	file, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = encDec.encode(file, p)
	err = errors.Join(err, file.Close())
	return err
}
