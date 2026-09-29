// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type SshSession interface {
	Output(cmd string) ([]byte, error)
	Run(cmd string) error
	Close() error
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
	SetStdin(stdin io.Reader)
	Shell() error
}

type sshSession struct {
	session *ssh.Session
}

func (s *sshSession) SetStdout(stdout io.Writer) {
	s.session.Stdout = stdout
}

func (s *sshSession) SetStderr(stderr io.Writer) {
	s.session.Stderr = stderr
}

func (s *sshSession) SetStdin(stdin io.Reader) {
	s.session.Stdin = stdin
}

func (s *sshSession) Shell() (err error) {

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		if e := term.Restore(fd, oldState); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to restore terminal state"), e)
		}
	}()

	// resize
	f, err := handleShellResize(s)
	if err != nil {
		return err
	}
	defer f()

	// terminal info
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
		height = 40
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,      // enable echoing
		ssh.TTY_OP_ISPEED: 115200, // input speed
		ssh.TTY_OP_OSPEED: 115200, // output speed
	}

	if err := s.session.RequestPty(termType, height, width, modes); err != nil {
		return err
	}

	if err := s.session.Shell(); err != nil {
		return err
	}

	if err := s.session.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *sshSession) Output(cmd string) ([]byte, error) {
	return s.session.Output(cmd)
}

func (s *sshSession) Run(cmd string) error {
	return s.session.Run(cmd)
}

func (s *sshSession) Close() error {
	return s.session.Close()
}
