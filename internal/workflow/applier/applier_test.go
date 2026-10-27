// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
)

var (
	// "proxy_ca_cert" is in both states with different definitions,
	// "registry_url" only in the current state and "dns_server" only in the target state.
	currentProxyCaCert = &state.Export{Node: "gateway", Capability: "org/squid-proxy", Path: "ca/cert"}
	targetProxyCaCert  = &state.Export{Cluster: "edge", Capability: "org/squid-proxy-ha", Path: "ca/pem"}
	registryURL        = &state.Export{Node: "gateway", Capability: "org/registry", Path: "url"}
	dnsServer          = &state.Export{Node: "resolver", Capability: "org/dns", Path: "address"}
)

// newSyncExportFixture returns a context of the stage. Both the current and the
// target state have the pre-stage "bootstrap" with exports.
func newSyncExportFixture(mode plan.Mode, stage string) *applyContext {
	current := state.NewState()
	current.Stages["bootstrap"] = &state.Stage{Export: state.Exports{
		"proxy_ca_cert": currentProxyCaCert.Clone(),
		"registry_url":  registryURL.Clone(),
	}}
	target := state.NewState()
	target.Stages["bootstrap"] = &state.Stage{Export: state.Exports{
		"proxy_ca_cert": targetProxyCaCert.Clone(),
		"dns_server":    dnsServer.Clone(),
	}}

	return &applyContext{
		stageStateManager: newStageStateManager(current, stage),
		plan:              &plan.Plan{Mode: mode},
		targetState:       target,
	}
}

func Test_syncExportData(t *testing.T) {
	tests := []struct {
		name string
		mode plan.Mode
		// modify the fixture before syncing
		setup func(ctx *applyContext)
		want  state.Exports
	}{
		{
			// I ◁ T
			name: "default: the target replaces and adds, the current is kept",
			mode: plan.ModeDefault,
			want: state.Exports{
				"proxy_ca_cert": targetProxyCaCert,
				"registry_url":  registryURL,
				"dns_server":    dnsServer,
			},
		},
		{
			name:  "default: the stage is new in the current state",
			mode:  plan.ModeDefault,
			setup: func(ctx *applyContext) { delete(ctx.state.Stages, "bootstrap") },
			want: state.Exports{
				"proxy_ca_cert": targetProxyCaCert,
				"dns_server":    dnsServer,
			},
		},
		{
			name:  "default: the stage is absent in the target state",
			mode:  plan.ModeDefault,
			setup: func(ctx *applyContext) { delete(ctx.targetState.Stages, "bootstrap") },
			want: state.Exports{
				"proxy_ca_cert": currentProxyCaCert,
				"registry_url":  registryURL,
			},
		},
		{
			// I|dom(T): the definition of the current state remains
			name: "prune: keeps the current definitions whose names are in the target",
			mode: plan.ModePrune,
			want: state.Exports{"proxy_ca_cert": currentProxyCaCert},
		},
		{
			name:  "prune: the stage is absent in the target state",
			mode:  plan.ModePrune,
			setup: func(ctx *applyContext) { delete(ctx.targetState.Stages, "bootstrap") },
			want:  nil,
		},
		{
			name: "destroy: removes all",
			mode: plan.ModeDestroy,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newSyncExportFixture(tt.mode, "bootstrap")
			if tt.setup != nil {
				tt.setup(ctx)
			}
			if err := syncExportData(ctx); err != nil {
				t.Fatalf("syncExportData() error = %v", err)
			}

			var got state.Exports
			if s, ok := ctx.State().FindStage("bootstrap"); ok {
				got = s.Export
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("exports = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("the recorded exports do not share entries with the target state", func(t *testing.T) {
		ctx := newSyncExportFixture(plan.ModeDefault, "bootstrap")
		if err := syncExportData(ctx); err != nil {
			t.Fatalf("syncExportData() error = %v", err)
		}
		ctx.targetState.Stages["bootstrap"].Export["dns_server"].Path = "changed"

		s, _ := ctx.State().FindStage("bootstrap")
		if got := s.Export["dns_server"].Path; got != dnsServer.Path {
			t.Errorf("recorded path = %q, want %q", got, dnsServer.Path)
		}
	})

	// _init and main have no exports, so nothing is recorded or saved
	for _, stage := range []string{state.InitStageName, state.MainStageName} {
		t.Run("skips the stage "+stage, func(t *testing.T) {
			ctx := newSyncExportFixture(plan.ModeDefault, stage)
			ctx.targetState.Export = state.Exports{"dns_server": dnsServer.Clone()}

			if err := syncExportData(ctx); err != nil {
				t.Fatalf("syncExportData() error = %v", err)
			}
			if ctx.State().Export != nil {
				t.Errorf("main exports = %v, want nil", ctx.State().Export)
			}
			if ctx.IsStateDirty() {
				t.Errorf("state must not be marked dirty")
			}
		})
	}
}
