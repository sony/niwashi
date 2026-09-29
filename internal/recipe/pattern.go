// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import "regexp"

const (
	versionPatternString          = "(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)"
	recipeIdPatternString         = "[a-z0-9][a-z0-9._-]*[a-z0-9]/[a-z0-9][a-z0-9._-]*[a-z0-9]"
	taskNamePatternString         = "[a-zA-Z0-9][a-zA-Z0-9 _.-]*"
	provideNamePatternString      = "[a-z0-9]([a-z0-9_-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9_-]*[a-z0-9])?)*"
	provideAttrKeyPatternString   = "[a-z][a-z0-9_]*"
	provideAttrValuePatternString = "[a-z0-9][a-z0-9_-]*"
	commandNamePatternString      = "[a-z][a-z0-9_-]*"
	fqidPatternString             = recipeIdPatternString + "@" + versionPatternString
)

// versionPattern matches a simplified semantic version of the form MAJOR.MINOR.PATCH.
// Pre-release and build metadata suffixes (e.g. "-alpha", "+001") are not accepted.
// Leading zeros in any numeric segment are rejected (e.g. "01.2.3" is invalid).
// Reference: https://semver.org/
var versionPattern = regexp.MustCompile("^" + versionPatternString + "$")

// recipeIdPattern matches a slash-separated identifier of the form "org/name".
// Each segment must start and end with a lowercase alphanumeric character.
// Dots, hyphens, and underscores are allowed within a segment.
// Uppercase letters and spaces are not permitted.
// Each segment must be at least two characters long.
// Examples: "my-org/my-recipe", "acme.corp/k8s-setup"
var recipeIdPattern = regexp.MustCompile("^" + recipeIdPatternString + "$")

// fqidPattern matches a fully-qualified recipe identifier of the form "org/name@MAJOR.MINOR.PATCH".
// It combines recipeIdPattern and versionPattern joined by "@".
// Example: "my-org/my-recipe@1.2.3"
var fqidPattern = regexp.MustCompile("^" + fqidPatternString + "$")

// taskNamePattern matches a human-readable task name.
// The name must start with an alphanumeric character (upper or lowercase).
// Subsequent characters may be alphanumeric, spaces, underscores, hyphens, or periods.
//
// Note: when used as a filename, the name undergoes the following conversions in order:
//  1. Converted to lowercase (Windows filesystems are case-insensitive, so "MyTask"
//     and "mytask" would refer to the same file on Windows).
//  2. Spaces and periods are replaced with filesystem-safe characters.
//
// Duplicate task name detection is performed after these conversions, so names that
// differ only in case, spaces, or periods are considered duplicates
// (e.g. "My Task", "my task", and "my.task" all collide).
var taskNamePattern = regexp.MustCompile("^" + taskNamePatternString + "$")

// provideNamePattern matches a dot-separated capability name.
// Each segment must start and end with a lowercase alphanumeric character.
// Hyphens and underscores are allowed within a segment; dots are used only as separators.
// A dot-separated hierarchy (e.g. "host.tool.ansible") is strongly recommended to
// keep the capability namespace organized and governable. Single-segment names
// (e.g. "external-instance") are technically valid but reserved for built-in system recipes.
// Examples: "host.tool.ansible", "node.capability.kubernetes", "external-instance"
var provideNamePattern = regexp.MustCompile("^" + provideNamePatternString + "$")

// provideAttrKeyPattern matches an attribute key used to disambiguate provides entries.
// Only lowercase alphanumeric characters and underscores are allowed, consistent with
// provideNamePattern. Dots and equals signs are excluded because they are structural
// separators in the alias string format (e.g. "host.tool.ansible.os=ubuntu").
var provideAttrKeyPattern = regexp.MustCompile("^" + provideAttrKeyPatternString + "$")

// provideAttrValuePattern matches an attribute value.
// Only lowercase alphanumeric characters, hyphens, and underscores are allowed,
// consistent with provideNamePattern and provideAttrKeyPattern.
// Dots are excluded because they are used as segment separators in the alias string,
// which means version strings like "2.16.3" cannot be used as values.
// Use FQID (org/name@MAJOR.MINOR.PATCH) to pin a specific version instead.
var provideAttrValuePattern = regexp.MustCompile("^" + provideAttrValuePatternString + "$")

// commandNamePattern matches a command name exposed by an adapter recipe.
// Follows POSIX utility naming conventions: lowercase letters, digits, hyphens,
// and underscores are allowed. The name must start with a lowercase letter.
// Uppercase letters are not permitted.
// Examples: "apply", "plan", "dry-run", "install_package"
var commandNamePattern = regexp.MustCompile("^" + commandNamePatternString + "$")
