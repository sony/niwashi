// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sony/niwashi/internal/file"
)

type Script struct {
	localPath   string
	executePath string
	firstLine   string
}

func NewScript(
	task TaskContext,
	fs file.FileSystem,
	src string,
	basepath string,
	te TemplateEngine,
	targetOS string) (*Script, error) {

	// transform
	script, err := te.Render(src)
	if err != nil {
		return nil, fmt.Errorf("failed to render script template for task %q: %w", TaskName(task), err)
	}

	// keep first line for shebang parsing
	var firstLine string
	if len(script) > 0 && script[0] == '#' {
		firstLine = strings.TrimSpace(script[:strings.Index(script, "\n")])
	}

	// adjust file extension based on the shebang
	ext := GetExtension(firstLine, targetOS)
	localPath := basepath + ext

	w, err := fs.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to open script file for writing: %s", localPath))
	}

	// Save script to a file under log dir
	_, err = io.WriteString(w, script)
	if err != nil {
		return nil, fmt.Errorf("failed to save script file %q: %w", localPath, err)
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close script file %q: %w", localPath, err)
	}

	return &Script{
		localPath: localPath,
		firstLine: firstLine,
	}, nil
}

func (s *Script) GetLocalPath() string {
	return s.localPath
}

func (s *Script) SetExecutePath(path string) {
	s.executePath = path
}

func (s *Script) GetArgs(os string) []string {
	// shbang + args + script path

	interp, args := ParseShebang(s.firstLine, os)

	cmd := []string{interp}
	cmd = append(cmd, args...)

	var script string
	if s.executePath != "" {
		script = s.executePath
	} else {
		script = s.localPath
	}

	if isPowerShell(script) {
		cmd = append(cmd, "-File")
	}
	cmd = append(cmd, script)

	return cmd
}

func isPowerShell(file string) bool {
	return strings.HasSuffix(strings.ToLower(file), ".ps1")
}

func ParseShebang(script, os string) (interp string, args []string) {
	interp, args = parseShebang(script)
	if interp == "" {
		interp = DefaultShell(os)
	}
	return interp, args
}

func parseShebang(script string) (interp string, args []string) {
	if len(script) < 3 || script[0:2] != "#!" {
		return "", nil
	}

	// get the first line of the script and trim CRLF
	line, _, _ := strings.Cut(script[2:], "\n")
	line = strings.TrimRight(line, "\r")

	// linux kernel limits
	if len(line) > 127 {
		line = line[:127]
	}

	// trim space
	line = strings.TrimLeft(line, " \t")

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}

	return fields[0], fields[1:]
}

func GetExtension(shebang, os string) string {

	interp, args := ParseShebang(shebang, os)
	if interp == "/usr/bin/env" && len(args) > 0 {
		interp = args[0]
	}

	// split interp with slash and get the last element
	parts := strings.Split(interp, "/")
	interp = parts[len(parts)-1]

	switch interp {
	case "sh", "bash":
		return ".sh"
	case "pwsh", "powershell":
		return ".ps1"
	default:
		return ""
	}
}

func DefaultShell(os string) string {
	if os == "windows" {
		return "powershell"
	}
	return "/bin/sh"
}
