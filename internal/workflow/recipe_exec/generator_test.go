// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
)

// HasWorkingTask must only report tasks that have been dispatched and not yet
// completed. A registered producer whose tasks were never produced must not
// count: after an error no new tasks are produced, so counting such a
// producer would keep the run state from ever reaching its exit hooks.
func TestTaskGenerator_HasWorkingTask(t *testing.T) {
	tests := []struct {
		name      string
		producers map[string]*Producer
		want      bool
	}{
		{
			name:      "no producers",
			producers: map[string]*Producer{},
			want:      false,
		},
		{
			name: "registered producer with only unproduced tasks",
			producers: map[string]*Producer{
				"pending": {filteredTasks: []*FilteredTask{{}}},
			},
			want: false,
		},
		{
			name: "a producer with a dispatched task",
			producers: map[string]*Producer{
				"pending": {filteredTasks: []*FilteredTask{{}}},
				"running": {workingTasks: []*Task{{}}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &taskGenerator{producers: tt.producers}
			if got := g.HasWorkingTask(); got != tt.want {
				t.Errorf("HasWorkingTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The hashes of the export values, as JSON (029 4):
//
//	printf '"PEM"' | sha256sum
//	printf '{"host":"gateway","port":3128}' | sha256sum
const (
	hashPEM   = "495d8f9c23c6e1682bfbd1e3570c6d8f85d799f00d4c43660e73921a86f7cf21"
	hashProxy = "b88e90edd7f1719ff480b1408a01009abcd8d08d4ae5a458430102316a94f231"
)

func TestCreateUserParams(t *testing.T) {
	// the param name and the export name differ so that a lookup of the
	// wrong one fails
	exportValues := map[string]any{
		"proxy_ca_cert": "PEM",
		"proxy":         map[string]any{"host": "gateway", "port": 3128},
	}
	job := func(params map[string]any, imports any) *workflow.Job {
		data := map[string]any{}
		if params != nil {
			data["params"] = params
		}
		if imports != nil {
			data["import"] = imports
		}
		return &workflow.Job{Data: data}
	}

	tests := []struct {
		name         string
		job          *workflow.Job
		want         map[string]any
		wantImported map[string]string // the hashes of the values actually injected, keyed by param name
		wantErr      string            // empty means no error
	}{
		{name: "nil data", job: &workflow.Job{}, want: map[string]any{}, wantImported: map[string]string{}},
		{
			name:         "params only",
			job:          job(map[string]any{"update": true}, nil),
			want:         map[string]any{"update": true},
			wantImported: map[string]string{},
		},
		{
			name:         "import is added to params",
			job:          job(map[string]any{"update": true}, types.Imports{"cert": {From: "proxy_ca_cert"}}),
			want:         map[string]any{"update": true, "cert": "PEM"},
			wantImported: map[string]string{"cert": hashPEM},
		},
		{
			name:         "subtree value",
			job:          job(nil, types.Imports{"proxy": {From: "proxy"}}),
			want:         map[string]any{"proxy": map[string]any{"host": "gateway", "port": 3128}},
			wantImported: map[string]string{"proxy": hashProxy},
		},
		{
			name:    "required export not found",
			job:     job(nil, types.Imports{"cert": {From: "proxy_endpoint"}}),
			wantErr: `param "cert": export "proxy_endpoint" not found`,
		},
		{
			name:    "required param also set by params",
			job:     job(map[string]any{"cert": "/etc/ca.pem"}, types.Imports{"cert": {From: "proxy_ca_cert"}}),
			wantErr: `param "cert" is set by both params and import (from "proxy_ca_cert")`,
		},
		{
			// the export name must not be looked up in params
			name:         "params has a key named after the export",
			job:          job(map[string]any{"proxy_ca_cert": "/etc/ca.pem"}, types.Imports{"cert": {From: "proxy_ca_cert"}}),
			want:         map[string]any{"proxy_ca_cert": "/etc/ca.pem", "cert": "PEM"},
			wantImported: map[string]string{"cert": hashPEM},
		},
		{
			// 029 3-3: no injection, params keeps its value; not recorded as imported (4)
			name:         "optional, export not found",
			job:          job(map[string]any{"cert": "/etc/ca.pem"}, types.Imports{"cert": {From: "proxy_endpoint", Optional: true}}),
			want:         map[string]any{"cert": "/etc/ca.pem"},
			wantImported: map[string]string{},
		},
		{
			// 029 3-3: the import overrides params when the export exists
			name:         "optional, export found",
			job:          job(map[string]any{"cert": "/etc/ca.pem"}, types.Imports{"cert": {From: "proxy_ca_cert", Optional: true}}),
			want:         map[string]any{"cert": "PEM"},
			wantImported: map[string]string{"cert": hashPEM},
		},
		{
			// the form read back from plan.json
			name: "import decoded from JSON",
			job: job(map[string]any{"update": true}, map[string]any{
				"cert": map[string]any{"from": "proxy_ca_cert"},
				"key":  map[string]any{"from": "proxy_key", "optional": true},
			}),
			want:         map[string]any{"update": true, "cert": "PEM"},
			wantImported: map[string]string{"cert": hashPEM},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotImported, err := createUserParams(tt.job, exportValues)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("createUserParams() error = %v", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("createUserParams() params = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(gotImported, tt.wantImported) {
					t.Errorf("createUserParams() imported = %v, want %v", gotImported, tt.wantImported)
				}
				return
			}
			if err == nil {
				t.Fatalf("createUserParams() = %v, want an error", got)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("createUserParams() error = %v, want %q in it", err, tt.wantErr)
			}
		})
	}

	t.Run("does not change the job data", func(t *testing.T) {
		params := map[string]any{"update": true}
		j := job(params, types.Imports{"cert": {From: "proxy_ca_cert"}})
		before, _ := json.Marshal(j.Data)

		if _, _, err := createUserParams(j, exportValues); err != nil {
			t.Fatalf("createUserParams() error = %v", err)
		}
		after, _ := json.Marshal(j.Data)
		if string(before) != string(after) {
			t.Errorf("job data changed: before %s, after %s", before, after)
		}
	})
}
