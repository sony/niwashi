// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"context"
	"fmt"
	"os"

	"github.com/sony/niwashi/internal/transport"
	"gopkg.in/yaml.v3"
)

type SimRemoteSession struct {
	Operations []any `yaml:"operations"`
}

type OpGetWorkRoot struct {
	Type string `yaml:"type"`
}

type OpUpload struct {
	Type   string `yaml:"type"`
	Local  string `yaml:"local"`
	Remote string `yaml:"remote"`
}

type OpDownload struct {
	Type   string       `yaml:"type"`
	Local  string       `yaml:"local"`
	Remote string       `yaml:"remote"`
	Perm   *os.FileMode `yaml:"perm,omitempty"`
}

type OpDownloadDir struct {
	Type      string `yaml:"type"`
	Local     string `yaml:"local"`
	Remote    string `yaml:"remote"`
	Recursive bool   `yaml:"recursive"`
}

type OpMkdirAll struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type OpExecute struct {
	Type       string            `yaml:"type"`
	Id         string            `yaml:"id"`
	LocalRoot  string            `yaml:"local_root"`
	RemoteRoot string            `yaml:"remote_root"`
	Cmd        []string          `yaml:"cmd"`
	Workdir    string            `yaml:"workdir"`
	Envs       map[string]string `yaml:"env"`
}

func NewSimRemoteSession(mode string) *SimRemoteSession {
	return &SimRemoteSession{}
}

func (s *SimRemoteSession) GetWorkRoot(ctx context.Context) (string, error) {
	s.Operations = append(s.Operations, &OpGetWorkRoot{Type: "get_work_root"})
	return "/tmp/sim_remote_session", nil
}

func (s *SimRemoteSession) Upload(ctx context.Context, local, remote string, perm *os.FileMode) error {
	s.Operations = append(s.Operations, &OpUpload{
		Type:   "upload",
		Local:  local,
		Remote: remote,
	})
	return nil
}

func (s *SimRemoteSession) Download(ctx context.Context, remote, local string, perm *os.FileMode) error {
	s.Operations = append(s.Operations, &OpDownload{
		Type:   "download",
		Local:  local,
		Remote: remote,
		Perm:   perm,
	})
	return nil
}

func (s *SimRemoteSession) DownloadDir(ctx context.Context, remote, local string, recursive bool) error {
	s.Operations = append(s.Operations, &OpDownloadDir{
		Type:      "download_dir",
		Local:     local,
		Remote:    remote,
		Recursive: recursive,
	})
	return nil
}

func (s *SimRemoteSession) MkdirAll(ctx context.Context, path string) error {
	s.Operations = append(s.Operations, &OpMkdirAll{
		Type: "mkdir_all",
		Path: path,
	})
	return nil
}

func (s *SimRemoteSession) Execute(ctx context.Context, exec transport.RemoteExecution) error {
	s.Operations = append(s.Operations, &OpExecute{
		Type:       "execute",
		Id:         exec.Id,
		Cmd:        exec.Cmd,
		Workdir:    exec.Workdir,
		Envs:       exec.Envs,
		LocalRoot:  exec.LocalRoot,
		RemoteRoot: exec.RemoteRoot,
	})
	return nil
}

func (s *SimRemoteSession) Close() error {
	v, err := yaml.Marshal(s)
	if err != nil {
		fmt.Println("Error marshaling SimulateCmd:", err)
	}
	fmt.Println("> Simulate remote action executed")
	fmt.Println(string(v))
	return nil
}
