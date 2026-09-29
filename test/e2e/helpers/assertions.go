// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package helpers

import (
	"encoding/json"
	"os"
	"strings"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// AssertFileExists asserts that a file exists
func AssertFileExists(t TestingT, path string) {
	t.Helper()
	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist
func AssertFileNotExists(t TestingT, path string) {
	t.Helper()
	if FileExists(path) {
		t.Errorf("expected file to not exist: %s", path)
	}
}

// AssertContains asserts that the haystack contains the needle
func AssertContains(t TestingT, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected to contain %q, but got:\n%s", needle, haystack)
	}
}

// AssertNotContains asserts that the haystack does not contain the needle
func AssertNotContains(t TestingT, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected to not contain %q, but got:\n%s", needle, haystack)
	}
}

// AssertJSONValid asserts that the string is valid JSON
func AssertJSONValid(t TestingT, jsonStr string) {
	t.Helper()
	var js interface{}
	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
		t.Errorf("expected valid JSON, but got error: %v\nContent:\n%s", err, jsonStr)
	}
}

// AssertJSONFileValid asserts that the file contains valid JSON
func AssertJSONFileValid(t TestingT, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	AssertJSONValid(t, string(data))
}

// AssertEqual asserts that two values are equal
func AssertEqual(t TestingT, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, but got %v", expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t TestingT, expected, actual interface{}) {
	t.Helper()
	if expected == actual {
		t.Errorf("expected values to be different, but both are %v", expected)
	}
}

// AssertNoError asserts that err is nil
func AssertNoError(t TestingT, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError asserts that err is not nil
func AssertError(t TestingT, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, but got nil")
	}
}
