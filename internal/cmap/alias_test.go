// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
)

func TestNewAlias(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *cmap.Alias
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cmap.NewAlias()
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlias_Add(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		key     string
		id      string
		version string
		attrs   map[string]string
	}{
		{
			name:    "basic test",
			key:     "tool:example",
			id:      "example-tool",
			version: "1.0.0",
			attrs: map[string]string{
				"os":   "linux",
				"arch": "amd64",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := cmap.NewAlias()
			a.Add(tt.key, tt.id, tt.version, tt.attrs)
		})
	}
}

func TestAlias_Find(t *testing.T) {

	testCases := []struct {
		name string
		// Named input parameters for target function.
		key       string
		insertKey func(a *cmap.Alias)
		// Expected output values.
		wantId      string
		wantVersion string
		wantErr     any
	}{
		{
			name:        "Find from empty Alias",
			key:         "nonexistent:key",
			wantId:      "",
			wantVersion: "",
			wantErr:     &cmap.NotFoundError{},
		},
		{
			name: "Key mismatch",
			key:  "nonexistent:key.os=linux.arch=amd64",
			insertKey: func(a *cmap.Alias) {
				a.Add("tool:example", "example-tool", "1.0.0", map[string]string{
					"os":   "linux",
					"arch": "amd64",
				})
			},
			wantId:      "",
			wantVersion: "",
			wantErr:     &cmap.NotFoundError{},
		},
		{
			name: "Find existing entry: complete match",
			key:  "tool:example.os=linux.arch=amd64",
			insertKey: func(a *cmap.Alias) {
				a.Add("tool:example", "example-tool", "1.0.0", map[string]string{
					"os":   "linux",
					"arch": "amd64",
				})
			},
			wantId:      "example-tool",
			wantVersion: "1.0.0",
			wantErr:     nil,
		},
		{
			name: "Find existing entry: complete match with multiple entries",
			key:  "tool:example.os=linux.arch=amd64",
			insertKey: func(a *cmap.Alias) {
				common_key := "tool:example"
				a.Add(common_key, "example-tool1", "1.0.0", map[string]string{
					"os":   "linux",
					"arch": "amd64",
				})
				a.Add(common_key, "example-tool2", "1.2.3", map[string]string{
					"os":   "linux",
					"arch": "aarch64",
				})
			},
			wantId:      "example-tool1",
			wantVersion: "1.0.0",
			wantErr:     nil,
		},
		{
			name: "Find existing entry: partial match with multiple entries",
			key:  "tool:example.arch=aarch64",
			insertKey: func(a *cmap.Alias) {
				common_key := "tool:example"
				a.Add(common_key, "example-tool1", "1.0.0", map[string]string{
					"os":   "linux",
					"arch": "amd64",
				})
				a.Add(common_key, "example-tool2", "1.2.3", map[string]string{
					"os":   "linux",
					"arch": "aarch64",
				})
			},
			wantId:      "example-tool2",
			wantVersion: "1.2.3",
			wantErr:     nil,
		},
		{
			name: "Find existing entry: multiple match",
			key:  "tool:example.os=linux",
			insertKey: func(a *cmap.Alias) {
				common_key := "tool:example"
				a.Add(common_key, "example-tool1", "1.0.0", map[string]string{
					"os":   "linux",
					"arch": "amd64",
				})
				a.Add(common_key, "example-tool2", "1.2.3", map[string]string{
					"os":   "linux",
					"arch": "aarch64",
				})
			},
			wantId:      "",
			wantVersion: "",
			wantErr:     &cmap.NotFoundError{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			a := cmap.NewAlias()
			if tt.insertKey != nil {
				tt.insertKey(a)
			}
			gotId, gotVersion, gotErr := a.Find(tt.key)
			if gotId != tt.wantId {
				t.Errorf("Find() gotId = %v, want %v", gotId, tt.wantId)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("Find() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			if tt.wantErr != nil {
				if !errors.As(gotErr, &tt.wantErr) {
					t.Errorf("Find() gotErr = %v, wantErr %v", gotErr, tt.wantErr)
				}
			}
		})
	}

}

func TestParseAlias(t *testing.T) {
	type args struct {
		fqid string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "basic test",
			args: args{
				fqid: "abc.key1=value1.key2=value2",
			},
			want: "abc",
			want1: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "no attrs1",
			args: args{
				fqid: "a",
			},
			want:    "a",
			want1:   map[string]string{},
			wantErr: false,
		},
		{
			name: "no attrs2",
			args: args{
				fqid: "abc.def.ghi",
			},
			want:    "abc.def.ghi",
			want1:   map[string]string{},
			wantErr: false,
		},
		{
			name: "attr normal",
			args: args{
				fqid: "my-adapter.key1=value1",
			},
			want: "my-adapter",
			want1: map[string]string{
				"key1": "value1",
			},
			wantErr: false,
		},
		{
			name: "error 1",
			args: args{
				fqid: ".",
			},
			want:    "",
			want1:   nil,
			wantErr: true,
			errMsg:  "empty segment found alias=\".\"",
		},
		{
			name: "error 2",
			args: args{
				fqid: "=",
			},
			want:    "",
			want1:   nil,
			wantErr: true,
			errMsg:  "attribute specified before identifier alias=\"=\"",
		},
		{
			name: "error 3",
			args: args{
				fqid: "a.=c=d",
			},
			want:    "",
			want1:   nil,
			wantErr: true,
			errMsg:  "empty attribute key alias=\"a.=c=d\"",
		},
		{
			name: "error 4",
			args: args{
				fqid: ".=",
			},
			want:    "",
			want1:   nil,
			wantErr: true,
			errMsg:  "empty segment found alias=\".=\"",
		},
		{
			name: "error 5",
			args: args{
				fqid: "abc.a=B.cde-sxe",
			},
			want:    "",
			want1:   nil,
			wantErr: true,
			errMsg:  "identifier found after attributes alias=\"abc.a=B.cde-sxe\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := cmap.ParseAlias(tt.args.fqid)
			if got != tt.want {
				t.Errorf("ParseAlias() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ParseAlias() got1 = %v, want %v", got1, tt.want1)
			}
			if (got2 != nil) != tt.wantErr {
				t.Errorf("ParseAlias() error = %v, wantErr %v", got2, tt.wantErr)
			}
			if got2 != nil && got2.Error() != tt.errMsg {
				t.Errorf("ParseAlias() error message = %v, want %v", got2.Error(), tt.errMsg)
			}
		})
	}
}
