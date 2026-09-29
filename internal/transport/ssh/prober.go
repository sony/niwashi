// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/types"
)

type Prober struct {
	t  *Transport
	fs file.FileSystem
}

func NewProber(t *Transport, fs file.FileSystem) *Prober {
	return &Prober{
		t:  t,
		fs: fs,
	}
}

func (p *Prober) GetSystemInfo(ctx context.Context) (_ *platform.SystemInfo, err error) {

	c, err := NewSshClient(ctx, p.t, p.fs)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := c.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()

	os, err := GetOsInfo(c)
	if err != nil {
		return nil, err
	}

	switch os {
	case "Linux": // "uname -s" returns "Linux" for Linux systems
		return GetLinuxInfo(c)
	// case platform.OsDarwin:
	// 	return GetDarwinInfo(s)
	default:
		if strings.Contains(os, "Windows") {
			return GetWindowsInfo(c)
		} else {
			return nil, fmt.Errorf("unsupported OS: %s", os)
		}
	}
}

func GetOsInfo(c SshClient) (_ string, err error) {

	s, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = s.Close()
	}()

	// Get OS name first
	osRaw, err := s.Output(`uname -s 2>/dev/null || powershell -NoProfile -Command "$PSVersionTable.OS"`)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(osRaw)), nil
}

func getUnameSm(c SshClient) (_ string, err error) {
	s, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = s.Close()
	}()

	os_arch, err := s.Output("uname -sm")
	if err != nil {
		return "", err
	}
	return string(os_arch), nil
}

func getOsRelease(c SshClient) (os_release string, err error) {
	s, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = s.Close()
	}()

	output, err := s.Output(". /etc/os-release && echo $ID $VERSION_ID $ID_LIKE")
	if err != nil {
		return "", err
	}
	os_release = string(output)
	return os_release, nil
}

func GetLinuxInfo(c SshClient) (*platform.SystemInfo, error) {
	os_arch, err := getUnameSm(c)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(os_arch))

	t := &platform.SystemInfo{
		Os:   fields[0],
		Arch: fields[1],
		Data: types.Dict{},
	}

	output, err := getOsRelease(c)
	if err != nil {
		return nil, err
	}

	fields = strings.Fields(string(output))

	t.Data["ID"] = fields[0]
	t.Data["VERSION_ID"] = fields[1]
	if len(fields) > 2 {
		t.Data["ID_LIKE"] = fields[2]
	}

	return t, nil
}

func GetWindowsInfo(c SshClient) (*platform.SystemInfo, error) {
	return nil, fmt.Errorf("not yet implemented GetWindowsInfo()")
}
