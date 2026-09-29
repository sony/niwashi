// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"maps"
	"slices"
	"testing"

	"github.com/sony/niwashi/internal/recipe"
)

type nilActionSpec struct{}

func (n *nilActionSpec) GetEnvTpl() map[string]string {
	return nil
}

func (n *nilActionSpec) Validate() error {
	return nil
}

func Test_recipeSpec_NewTaskSpec(t *testing.T) {

	t.Run("normal case", func(t *testing.T) {
		nas := &nilActionSpec{}
		sc := map[string]*recipe.StateChangeOperation{
			"state1": {
				Op:    "set",
				Value: "value1",
			},
		}
		src := &recipe.Task{
			Name: "test_task",
			Action: recipe.Action{
				ActionType: "dummy",
				Spec:       nas,
			},
			Where:        "test_where",
			StateChanges: sc,
		}

		r := &recipeSpec{}
		got := r.NewTaskSpec(src)

		// from action.TaskSpec
		// GetName() string
		// GetRecipe() RecipeSpec
		// GetActionType() string
		// GetActionSpec() ActionSpec
		if got.GetName() != src.Name {
			t.Errorf("Expected name %s, got %s", src.Name, got.GetName())
		}
		if got.GetRecipe() != r {
			t.Errorf("Expected recipe %v, got %v", r, got.GetRecipe())
		}
		if got.GetActionType() != "dummy" {
			t.Errorf("Expected action type 'dummy', got %s", got.GetActionType())
		}
		if got.GetActionSpec() != nas {
			t.Errorf("Expected action spec %v, got %v", nas, got.GetActionSpec())
		}

		// from TaskSpec in recipe_exec
		// GetWhere() string
		// GetStateChanges() map[string]*recipe.StateChangeOperation
		if got.GetWhere() != src.Where {
			t.Errorf("Expected where %s, got %s", src.Where, got.GetWhere())
		}

		if !maps.Equal(got.GetStateChanges(), sc) {
			t.Errorf("Expected state changes %v, got %v", sc, got.GetStateChanges())
		}
	})

	t.Run("adapter case", func(t *testing.T) {

		argvTpl := []string{
			"hello",
			"world",
		}
		workDirTpl := "./testdata/workspace"
		inheritEnv := []string{
			"HOME",
			"USER",
		}

		src := &recipe.Task{
			Name: "test_task",
			Action: recipe.Action{
				ActionType: recipe.ToolRunActionType,
				Spec: &recipe.ToolRunSpec{
					ArgvTemplate:    argvTpl,
					WorkdirTemplate: workDirTpl,
					InheritEnv:      inheritEnv,
				},
			},
		}

		r := &recipeSpec{}
		got := r.NewTaskSpec(src)

		got2, ok := got.(CallerTaskSpec)
		if !ok {
			t.Fatalf("Expected CallerTaskSpec, got %T", got)
		}

		// from action.CallerTaskSpec
		// GetArgvTpl() []string
		// GetWorkDirTpl() string
		// GetInheritEnv() []string
		if !slices.Equal(got2.GetArgvTpl(), argvTpl) {
			t.Errorf("Expected argv template %v, got %v", argvTpl, got2.GetArgvTpl())
		}
		if got2.GetWorkDirTpl() != workDirTpl {
			t.Errorf("Expected work dir template %s, got %s", workDirTpl, got2.GetWorkDirTpl())
		}
		if !slices.Equal(got2.GetInheritEnv(), inheritEnv) {
			t.Errorf("Expected inherit env %v, got %v", inheritEnv, got2.GetInheritEnv())
		}
	})

}
