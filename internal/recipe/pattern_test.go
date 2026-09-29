// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"regexp"
	"testing"
)

func TestRecipePattern_Validate(t *testing.T) {
	type testCase struct {
		input  string
		result bool
	}
	tests := []struct {
		name      string // description of this test case
		re        *regexp.Regexp
		testCases []testCase
	}{
		{
			name: "version pattern",
			re:   versionPattern,
			testCases: []testCase{
				{input: "1.0.0", result: true},
				{input: "v1.0.0", result: false},
				{input: "1.0.0-alpha", result: false},
				{input: "1.0", result: false},
				{input: "1", result: false},
				{input: "1.0.02", result: false},
			},
		},
		{
			name: "recipe id pattern",
			re:   recipeIdPattern,
			testCases: []testCase{
				{input: "org/my-recipe", result: true},
				{input: "org/my-recipe.xyz_xyz", result: true},
				{input: "org/my-recipe.xyz_xyz.abc-abc", result: true},
				{input: "01/my-recipe", result: true},
				{input: "my-recipe", result: false},      // no org
				{input: "a_/my-recipe", result: false},   // org name ends with _
				{input: "a/my-recipe", result: false},    // short org name
				{input: "org/my-recipe_", result: false}, // recipe name ends with _
			},
		},
		{
			name: "FQID pattern",
			re:   fqidPattern,
			testCases: []testCase{
				{input: "org/my-recipe@1.0.0", result: true},
				{input: "org/my-recipe@a.b.c", result: false},
			},
		},
		{
			name: "task name pattern",
			re:   taskNamePattern,
			testCases: []testCase{
				{input: "task1", result: true},
				{input: "Task2", result: true},
				{input: "a", result: true},
				{input: "a.b.c", result: true},
				{input: "a_b_c", result: true},
				{input: "a-b-c", result: true},
				{input: "a b c", result: true},
				{input: "", result: false},
				{input: "invalid/task/name", result: false},
			},
		},
		{
			name: "provide name pattern",
			re:   provideNamePattern,
			testCases: []testCase{
				{input: "provide1", result: true},
				{input: "provide-1", result: true},
				{input: "a", result: true},
				{input: "0", result: true},
				{input: "9", result: true},
				{input: "provide.1", result: true},
				{input: "a.b.c", result: true},
				{input: "a.b.012", result: true},
				{input: "", result: false},
			},
		},
		{
			name: "provide attr key pattern",
			re:   provideAttrKeyPattern,
			testCases: []testCase{
				{input: "a", result: true},
				{input: "attr1", result: true},
				{input: "attr_1", result: true},

				{input: "1attr", result: false},
				{input: "attr.1", result: false},
				{input: "attr-1", result: false},
				{input: "attr 1", result: false},
			},
		},
		{
			name: "provide attr value pattern",
			re:   provideAttrValuePattern,
			testCases: []testCase{
				{input: "value1", result: true},
				{input: "value-1", result: true},
				{input: "value_1", result: true},
				{input: "0", result: true},
				{input: "9", result: true},
			},
		},
		{
			name: "command name pattern",
			re:   commandNamePattern,
			testCases: []testCase{
				{input: "cmd1", result: true},
				{input: "cmd-1", result: true},
				{input: "cmd_1", result: true},
				{input: "a", result: true},

				{input: "1_cmd", result: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tc := range tt.testCases {
				result := tt.re.MatchString(tc.input)
				if result != tc.result {
					t.Errorf("MatchString(%q) = %v, want %v", tc.input, result, tc.result)
				}
			}
		})
	}
}
