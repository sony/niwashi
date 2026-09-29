// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
)

type KnownHosts struct {
	Entries []string
	FS      file.FileSystem
}

func NewKnownHosts(fs file.FileSystem) *KnownHosts {
	return &KnownHosts{
		Entries: []string{},
		FS:      fs,
	}
}

func NewKnownHostsFromFile(fs file.FileSystem, path string) (*KnownHosts, error) {
	kh := NewKnownHosts(fs)
	if err := kh.LoadHostEntries(path); err != nil {
		return nil, err
	}
	return kh, nil
}

func (k *KnownHosts) LoadHostEntries(knownHostsPath string) error {
	f, err := k.FS.Open(knownHostsPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k.Entries = append(k.Entries, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (k *KnownHosts) CopyHostEntry(c *Transport) error {

	if c == nil {
		return fmt.Errorf("connection is nil")
	}
	if c.HostKey.Method != "" && c.HostKey.Method != HostKeyMethodKnownHostsFile {
		return fmt.Errorf("unsupported host key method: %s", c.HostKey.Method)
	}
	if c.HostKey.KnownHostsPath == "" {
		return fmt.Errorf("knownHostsPath is empty")
	}

	f, err := k.FS.Open(c.HostKey.KnownHostsPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	patterns := []string{
		c.Address.Host,
		fmt.Sprintf("[%s]:%d", c.Address.Host, c.Address.Port),
	}

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		for _, pattern := range patterns {
			if strings.HasPrefix(line, pattern+" ") || strings.HasPrefix(line, pattern+",") {
				lines = append(lines, line)
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(lines) == 0 {
		logger.Warn("No matching host key entry found in known_hosts for %s", c.Address.Host)
		return nil
	}

	k.Entries = append(k.Entries, lines...)

	return nil
}

func (k *KnownHosts) RemoveHostEntry(t *Transport) error {

	if t == nil {
		return fmt.Errorf("connection is nil")
	}
	if t.HostKey.Method != "" && t.HostKey.Method != HostKeyMethodKnownHostsFile {
		return fmt.Errorf("unsupported host key method: %s", t.HostKey.Method)
	}
	if t.HostKey.KnownHostsPath == "" {
		return fmt.Errorf("knownHostsPath is empty")
	}

	patterns := []string{
		t.Address.Host,
		fmt.Sprintf("[%s]:%d", t.Address.Host, t.Address.Port),
	}

	var filteredLines []string
	for _, entry := range k.Entries {
		line := strings.TrimSpace(entry)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var found bool
		for _, pattern := range patterns {
			if strings.HasPrefix(line, pattern+" ") || strings.HasPrefix(line, pattern+",") {
				found = true
				break
			}
		}

		if !found {
			filteredLines = append(filteredLines, line)
		}
	}

	// replace entries with filtered lines
	k.Entries = filteredLines
	return nil
}

func (k *KnownHosts) SaveToFile(path string, createSafeDir bool) (err error) {
	if createSafeDir {
		dir := filepath.Dir(path)
		if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
			return err
		}
	}

	f, err := k.FS.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	writer := bufio.NewWriter(f)
	for _, entry := range k.Entries {
		if _, err := fmt.Fprintln(writer, entry); err != nil {
			return err
		}
	}
	return writer.Flush()
}
