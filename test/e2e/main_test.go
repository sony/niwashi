// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var testBinaryPath string

func TestMain(m *testing.M) {
	path, err := prepareBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare nwsctl binary: %v\n", err)
		os.Exit(1)
	}
	testBinaryPath = path
	os.Exit(m.Run())
}

// prepareBinary returns the path to the nwsctl binary.
// If NWSCTL_BINARY is set, that path is used directly (pre-built binary).
// Otherwise, the binary is built from source.
func prepareBinary() (string, error) {
	if path := os.Getenv("NWSCTL_BINARY"); path != "" {
		return path, nil
	}
	return buildBinary()
}

func buildBinary() (string, error) {
	root, err := findProjectRoot()
	if err != nil {
		return "", err
	}

	outputPath := filepath.Join(root, "bin", "nwsctl-e2e-test")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/nwsctl")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}
	return outputPath, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
