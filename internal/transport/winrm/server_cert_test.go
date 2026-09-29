// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

import (
	"reflect"
	"testing"
)

func TestServerCert_Clone(t *testing.T) {
	tests := []struct {
		name string
		arg  *ServerCert
		want *ServerCert
	}{
		{name: "nil ServerCert", arg: nil, want: nil},
		{
			name: "with values",
			arg:  &ServerCert{InsecureSkipVerify: true, CACertPath: "/ca.pem"},
			want: &ServerCert{InsecureSkipVerify: true, CACertPath: "/ca.pem"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.Clone()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestServerCert_Merge(t *testing.T) {
	tests := []struct {
		name  string
		base  *ServerCert
		other *ServerCert
		want  *ServerCert
	}{
		{
			name:  "other's CACertPath takes precedence",
			base:  &ServerCert{CACertPath: "/base-ca.pem"},
			other: &ServerCert{CACertPath: "/other-ca.pem"},
			want:  &ServerCert{CACertPath: "/other-ca.pem"},
		},
		{
			name:  "empty other keeps base's CACertPath",
			base:  &ServerCert{CACertPath: "/base-ca.pem"},
			other: &ServerCert{},
			want:  &ServerCert{CACertPath: "/base-ca.pem"},
		},
		{
			name:  "InsecureSkipVerify is OR'd",
			base:  &ServerCert{InsecureSkipVerify: false},
			other: &ServerCert{InsecureSkipVerify: true},
			want:  &ServerCert{InsecureSkipVerify: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.base.Merge(tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
