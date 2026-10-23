// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
)

// The param names and the export names differ in every case below,
// so that a check looking up the wrong one fails.
const (
	importCapId      = "org/ca-trust"
	importParam      = "cert"          // param name of the recipe
	importExportName = "proxy_ca_cert" // export name of the preceding stage
)

// importStage returns a stage whose node "cp" has the capability with the
// given params and imports.
func importStage(params types.Params, imports types.Imports) *state.Stage {
	s := &state.Stage{Inventory: state.NewInventory()}
	s.Inventory.Nodes["cp"] = &node.Node{Capabilities: cap.CapabilityList{
		importCapId: {Params: params, Import: imports},
	}}
	return s
}

// generatorImportStage returns a stage whose generator "vm" has the given
// params and imports.
func generatorImportStage(params types.Params, imports types.Imports) *state.Stage {
	return &state.Stage{
		Inventory: state.NewInventory(),
		Infrastructure: &state.Infrastructure{Generators: map[string]*instance.Generator{
			"vm": {Provisioner: "org/vm", Params: params, Import: imports},
		}},
	}
}

func importOf(from string, optional bool) types.Imports {
	return types.Imports{importParam: {From: from, Optional: optional}}
}

func importExports() state.Exports {
	return state.Exports{importExportName: {Node: "gateway", Capability: "org/squid-proxy", Path: "ca/cert"}}
}

func TestStage_ValidateImports(t *testing.T) {
	tests := []struct {
		name    string
		stage   *state.Stage
		exports state.Exports
		wantErr string // empty means no error
	}{
		{
			name:    "export exists",
			stage:   importStage(nil, importOf(importExportName, false)),
			exports: importExports(),
		},
		{
			name:    "export not found",
			stage:   importStage(nil, importOf(importExportName, false)),
			exports: state.Exports{},
			wantErr: `export "proxy_ca_cert" not found`,
		},
		{
			// the param name must not be looked up as an export name
			name:    "only an export named after the param",
			stage:   importStage(nil, importOf(importExportName, false)),
			exports: state.Exports{importParam: {Node: "gateway", Capability: "org/squid-proxy", Path: "ca/cert"}},
			wantErr: `export "proxy_ca_cert" not found`,
		},
		{
			name:    "optional, export not found",
			stage:   importStage(nil, importOf(importExportName, true)),
			exports: state.Exports{},
		},
		{
			name:    "param also set by params",
			stage:   importStage(types.Params{importParam: "/etc/ca.pem"}, importOf(importExportName, false)),
			exports: importExports(),
			wantErr: `param "cert" is set by both params and import`,
		},
		{
			// the export name must not be looked up in params
			name:    "params has a key named after the export",
			stage:   importStage(types.Params{importExportName: "/etc/ca.pem"}, importOf(importExportName, false)),
			exports: importExports(),
		},
		{
			// params is the fallback for an optional import (029 3-3)
			name:    "optional, param also set by params",
			stage:   importStage(types.Params{importParam: "/etc/ca.pem"}, importOf(importExportName, true)),
			exports: importExports(),
		},
		{
			name: "one of the imports not found",
			stage: importStage(nil, types.Imports{
				importParam: {From: importExportName},
				"endpoint":  {From: "proxy_endpoint"},
			}),
			exports: importExports(),
			wantErr: `export "proxy_endpoint" not found`,
		},
		{
			name: "cluster capability, export not found",
			stage: func() *state.Stage {
				s := &state.Stage{Inventory: state.NewInventory()}
				s.Inventory.Clusters["k8s"] = &cluster.Cluster{Capabilities: cap.CapabilityList{
					importCapId: {Import: importOf(importExportName, false)},
				}}
				return s
			}(),
			exports: state.Exports{},
			wantErr: `cluster "k8s"`,
		},
		{
			name:    "no import",
			stage:   importStage(types.Params{importParam: "/etc/ca.pem"}, nil),
			exports: state.Exports{},
		},
		{
			name:    "generator, export exists",
			stage:   generatorImportStage(nil, importOf(importExportName, false)),
			exports: importExports(),
		},
		{
			name:    "generator, export not found",
			stage:   generatorImportStage(nil, importOf(importExportName, false)),
			exports: state.Exports{},
			wantErr: `generator "vm": param "cert": export "proxy_ca_cert" not found`,
		},
		{
			name:    "generator, param also set by params",
			stage:   generatorImportStage(types.Params{importParam: "/etc/ca.pem"}, importOf(importExportName, false)),
			exports: importExports(),
			wantErr: `generator "vm": param "cert" is set by both params and import`,
		},
		{
			name:    "generator, params has a key named after the export",
			stage:   generatorImportStage(types.Params{importExportName: "/etc/ca.pem"}, importOf(importExportName, false)),
			exports: importExports(),
		},
		{
			name:    "generator, optional, export not found",
			stage:   generatorImportStage(types.Params{importParam: "/etc/ca.pem"}, importOf(importExportName, true)),
			exports: state.Exports{},
		},
		{
			name:    "nil stage",
			stage:   nil,
			exports: state.Exports{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stage.ValidateImports(tt.exports)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateImports() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateImports() succeeded unexpectedly")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateImports() error = %v, want %q in it", err, tt.wantErr)
			}
		})
	}
}

func TestStage_ValidateImports_ErrorMessage(t *testing.T) {
	err := importStage(types.Params{importParam: "/etc/ca.pem"}, importOf(importExportName, false)).
		ValidateImports(importExports())
	if err == nil {
		t.Fatal("ValidateImports() succeeded unexpectedly")
	}
	// 029 3-3: the message has the subject, the capability, the param and the export
	for _, want := range []string{`node "cp"`, `capability "org/ca-trust"`, `param "cert"`, `"proxy_ca_cert"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ValidateImports() error = %v, want %s in it", err, want)
		}
	}
}
