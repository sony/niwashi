package recipe_exec_test

import (
	"errors"
	"maps"
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/applier"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

type fakeStateManager struct {
	patches []patch.Patch
	err     error
}

func (f *fakeStateManager) ApplyPatches(patches []patch.Patch, _ patch.PathConstraint, _ func(*state.Accessor) error) error {
	f.patches = append(f.patches, patches...)
	return f.err
}

func (f *fakeStateManager) State() *state.Accessor { return nil }
func (f *fakeStateManager) SaveState() error       { return nil }
func (f *fakeStateManager) IsStateDirty() bool     { return false }

func newTestRecipeSpec(id, version string) recipe_exec.RecipeSpec {
	r := &recipe.Recipe{
		Kind:     recipe.KindNode,
		Metadata: &recipe.Metadata{Id: id, Version: version},
		Spec:     &recipe.Spec{},
	}
	r.SetDefaults()
	return recipe_exec.NewRecipeSpec(r, recipe.OperationConstruct)
}

func TestAddCapability(t *testing.T) {
	const (
		prefix  = "/inventory/nodes"
		target  = "node-1"
		version = "1.0.0"
	)

	tests := []struct {
		name       string
		recipeId   string
		jobData    any
		applyErr   error
		wantParams types.Params
	}{
		{
			// job.Data always arrives here after a round-trip through plan.json
			// (nwsctl plan writes it, nwsctl apply reads it back), so by the time
			// AddCapability runs, "params" is decoded as a plain map[string]any —
			// never as types.Params. This is the realistic success case.
			name:       "params present as plain map[string]any",
			recipeId:   "my/recipe",
			jobData:    map[string]any{"params": map[string]any{"memory": "4096"}},
			wantParams: types.Params{"memory": "4096"},
		},
		{
			// "データが空": the params entry exists and is the correct type, but holds
			// no keys. This must still be treated as a valid (empty) Params value,
			// not fall through to the default.
			name:       "params present but empty",
			recipeId:   "my/recipe",
			jobData:    map[string]any{"params": map[string]any{}},
			wantParams: types.Params{},
		},
		{
			// types.Params/types.Dict is a defined type over map[string]any, not an
			// alias, so it does NOT satisfy the `p.(map[string]any)` assertion even
			// though the underlying shape is identical. In practice this value never
			// occurs post-JSON round-trip, but the fallback must still be safe.
			name:       "params present as types.Params falls back to default",
			recipeId:   "my/recipe",
			jobData:    map[string]any{"params": types.Params{"memory": "4096"}},
			wantParams: types.Params{},
		},
		{
			name:       "params present as unrelated type falls back to default",
			recipeId:   "my/recipe",
			jobData:    map[string]any{"params": "not-a-params-map"},
			wantParams: types.Params{},
		},
		{
			name:       "params key missing from job data",
			recipeId:   "my/recipe",
			jobData:    map[string]any{"other": "value"},
			wantParams: types.Params{},
		},
		{
			name:       "job data is not a map at all",
			recipeId:   "my/recipe",
			jobData:    "not-a-map",
			wantParams: types.Params{},
		},
		{
			name:       "job data is nil",
			recipeId:   "my/recipe",
			jobData:    nil,
			wantParams: types.Params{},
		},
		{
			name:       "recipe id containing a slash is escaped in the patch path",
			recipeId:   "org/team.recipe",
			jobData:    map[string]any{"params": map[string]any{"a": 1}},
			wantParams: types.Params{"a": 1},
		},
		{
			// The ApplyPatches error must be returned to the caller (wrapping the
			// original), and the capability patch must still have been built and
			// submitted before the error is handled.
			name:       "ApplyPatches returns an error",
			recipeId:   "my/recipe",
			jobData:    map[string]any{"params": map[string]any{"a": 1}},
			applyErr:   errors.New("boom"),
			wantParams: types.Params{"a": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &fakeStateManager{err: tt.applyErr}
			r := newTestRecipeSpec(tt.recipeId, version)
			job := &workflow.Job{Data: tt.jobData}

			err := recipe_exec.AddCapability(prefix, sm, r, target, job)
			if !errors.Is(err, tt.applyErr) {
				t.Errorf("AddCapability() error = %v, want %v", err, tt.applyErr)
			}

			if len(sm.patches) != 1 {
				t.Fatalf("expected exactly 1 patch to be applied, got %d", len(sm.patches))
			}

			wantPath := recipe_exec.CapabilityRoot(prefix, r, target)
			if got := sm.patches[0].GetPath(); got != wantPath {
				t.Errorf("path = %q, want %q", got, wantPath)
			}

			ap, ok := sm.patches[0].(*patch.AddPatch)
			if !ok {
				t.Fatalf("expected *patch.AddPatch, got %T", sm.patches[0])
			}
			c, ok := ap.Value.(*cap.Capability)
			if !ok {
				t.Fatalf("expected *cap.Capability value, got %T", ap.Value)
			}

			if c.Version != version {
				t.Errorf("Version = %q, want %q", c.Version, version)
			}
			if !maps.Equal(c.Params, tt.wantParams) {
				t.Errorf("Params = %#v, want %#v", c.Params, tt.wantParams)
			}
		})
	}
}

func newExistingCapability() *cap.Capability {
	return &cap.Capability{
		Version: "1.0.0",
		Params:  types.Params{"memory": "1024"},
		Store:   types.Dict{"ip": "10.0.0.1"},
	}
}

func nodeStateWith(c *cap.Capability) *state.State {
	caps := cap.CapabilityList{}
	if c != nil {
		caps["my/recipe"] = c
	}
	return &state.State{Inventory: &state.Inventory{
		Nodes: map[string]*node.Node{"node-1": {Capabilities: caps}},
	}}
}

func clusterStateWith(c *cap.Capability) *state.State {
	caps := cap.CapabilityList{}
	if c != nil {
		caps["my/recipe"] = c
	}
	return &state.State{Inventory: &state.Inventory{
		Clusters: map[string]*cluster.Cluster{"cluster-1": {Capabilities: caps}},
	}}
}

func lookupCapability(s *state.Accessor, prefix, target, id string) *cap.Capability {
	switch prefix {
	case "/inventory/nodes":
		if n, ok := s.Inventory.Nodes[target]; ok {
			return n.Capabilities[id]
		}
	case "/inventory/clusters":
		if c, ok := s.Inventory.Clusters[target]; ok {
			return c.Capabilities[id]
		}
	}
	return nil
}

// TestUpdateCapability applies the patches to a real state (via the
// applier's state manager) instead of only inspecting them, because the
// failure modes of UpdateCapability (an unsupported path, or a value type
// that Capability.Add rejects) only surface when the patch is applied.
func TestUpdateCapability(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		target  string
		initial *state.State
		jobData any
		want    *cap.Capability
		wantErr bool
	}{
		{
			name:    "node: version and params updated, store kept",
			prefix:  "/inventory/nodes",
			target:  "node-1",
			initial: nodeStateWith(newExistingCapability()),
			jobData: map[string]any{"params": map[string]any{"memory": "2048"}},
			want: &cap.Capability{
				Version: "1.1.0",
				Params:  types.Params{"memory": "2048"},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
		},
		{
			name:    "cluster: version and params updated, store kept",
			prefix:  "/inventory/clusters",
			target:  "cluster-1",
			initial: clusterStateWith(newExistingCapability()),
			jobData: map[string]any{"params": map[string]any{"memory": "2048"}},
			want: &cap.Capability{
				Version: "1.1.0",
				Params:  types.Params{"memory": "2048"},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
		},
		{
			// Same default as AddCapability: without params in the job data,
			// the recorded params are replaced with an empty map.
			name:    "params missing from job data are recorded as empty",
			prefix:  "/inventory/nodes",
			target:  "node-1",
			initial: nodeStateWith(newExistingCapability()),
			jobData: map[string]any{},
			want: &cap.Capability{
				Version: "1.1.0",
				Params:  types.Params{},
				Store:   types.Dict{"ip": "10.0.0.1"},
			},
		},
		{
			// The patch fails because there is no capability to update: the
			// error is returned and the state stays unchanged.
			name:    "capability absent from state returns an error and leaves state unchanged",
			prefix:  "/inventory/nodes",
			target:  "node-1",
			initial: nodeStateWith(nil),
			jobData: map[string]any{"params": map[string]any{"memory": "2048"}},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := applier.NewStateManager(state.NewAccessor(tt.initial))
			r := newTestRecipeSpec("my/recipe", "1.1.0")
			job := &workflow.Job{Data: tt.jobData}

			err := recipe_exec.UpdateCapability(tt.prefix, sm, r, tt.target, job)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCapability() error = %v, wantErr %v", err, tt.wantErr)
			}

			got := lookupCapability(sm.State(), tt.prefix, tt.target, "my/recipe")
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("capability after update = %#v, want %#v", got, tt.want)
			}
		})
	}
}
