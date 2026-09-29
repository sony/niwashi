// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/sony/niwashi/test/e2e/helpers"
)

func TestVersion(t *testing.T) {
	cli := helpers.NewCLI(testBinaryPath)
	stdout, _, err := cli.Run("version")
	if err != nil {
		t.Fatalf("nwsctl version failed: %v", err)
	}
	if !strings.Contains(stdout, "nwsctl version") {
		t.Errorf("expected output to contain 'nwsctl version', got:\n%s", stdout)
	}
}
