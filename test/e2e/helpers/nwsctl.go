// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TestingT is a minimal interface for testing.T
type TestingT interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Logf(format string, args ...interface{})
	Helper()
	FailNow()
}

// CLI wraps nwsctl command execution
type CLI struct {
	binaryPath string
	workDir    string
	env        []string
}

// NewCLI creates a new CLI wrapper
func NewCLI(binaryPath string) *CLI {
	return &CLI{
		binaryPath: binaryPath,
		env:        os.Environ(),
	}
}

// WithWorkDir sets the working directory for commands
func (c *CLI) WithWorkDir(dir string) *CLI {
	c.workDir = dir
	return c
}

// WithEnv adds environment variables
func (c *CLI) WithEnv(key, value string) *CLI {
	c.env = append(c.env, fmt.Sprintf("%s=%s", key, value))
	return c
}

// Run executes a nwsctl command and returns stdout, stderr, and error
func (c *CLI) Run(args ...string) (string, string, error) {
	cmd := exec.Command(c.binaryPath, args...)
	cmd.Dir = c.workDir
	cmd.Env = c.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// RunSuccess executes a command and fails the test if it returns an error
func (c *CLI) RunSuccess(t TestingT, args ...string) (stdout, stderr string) {
	t.Helper()
	stdout, stderr, err := c.Run(args...)
	if err != nil {
		// Format error message with clear sections
		t.Errorf("Command failed: %v", err)
		t.Errorf("Command: %s", c.CommandString(args...))
		t.Errorf("Working directory: %s", c.workDir)
		t.Errorf("\n=== STDOUT ===\n%s\n=== END STDOUT ===", stdout)
		t.Errorf("\n=== STDERR ===\n%s\n=== END STDERR ===", stderr)
		t.FailNow()
	}
	return stdout, stderr
}

// RunExpectError executes a command and expects it to fail
func (c *CLI) RunExpectError(t TestingT, args ...string) (stdout, stderr string) {
	t.Helper()
	stdout, stderr, err := c.Run(args...)
	if err == nil {
		t.Fatalf("command unexpectedly succeeded\nArgs: %v\nStdout: %s",
			args, stdout)
	}
	return stdout, stderr
}

// CommandString returns the command string for debugging
func (c *CLI) CommandString(args ...string) string {
	parts := []string{c.binaryPath}
	parts = append(parts, args...)
	return strings.Join(parts, " ")
}

// RunWithOutput executes a command with real-time output to test log
// Useful for debugging or when you want to see output as it happens
func (c *CLI) RunWithOutput(t TestingT, args ...string) error {
	t.Helper()
	t.Logf("Running: %s", c.CommandString(args...))
	t.Logf("Working directory: %s", c.workDir)

	cmd := exec.Command(c.binaryPath, args...)
	cmd.Dir = c.workDir
	cmd.Env = c.env

	// Create buffers to capture output
	var stdout, stderr bytes.Buffer

	// Use io.MultiWriter to write to both buffer and test log
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Log the output
	if stdout.Len() > 0 {
		t.Logf("STDOUT:\n%s", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Logf("STDERR:\n%s", stderr.String())
	}

	if err != nil {
		t.Logf("Command failed with error: %v", err)
	}

	return err
}

// RunSuccessWithOutput is like RunSuccess but logs output in real-time
func (c *CLI) RunSuccessWithOutput(t TestingT, args ...string) {
	t.Helper()
	err := c.RunWithOutput(t, args...)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
}
