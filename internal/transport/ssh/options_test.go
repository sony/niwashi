// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"reflect"
	"testing"
)

func Test_getEnvOptions(t *testing.T) {
	tests := []struct {
		name    string            // description of this test case
		envVars map[string]string // environment variables to set for this test case
		want    *Options
	}{
		{
			name: "no environment variables set",
			want: &Options{},
		},
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"NWS_SSH_CONNECT_TIMEOUT_SEC":   "99",
				"NWS_SSH_HANDSHAKE_TIMEOUT_SEC": "88",
				"NWS_SSH_RETRY_MAX_COUNT":       "77",
				"NWS_SSH_RETRY_INTERVAL_SEC":    "66",
			},
			want: &Options{
				ConnectTimeoutSec:   99,
				HandshakeTimeoutSec: 88,
				RetryMaxCount:       77,
				RetryIntervalSec:    66,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			got := getEnvOptions()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEnvOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_SetDefaults(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		base *Options
		want *Options
	}{
		{
			name: "set defaults",
			base: &Options{},
			want: defaultOptions,
		},
		{
			name: "already has values",
			base: &Options{
				ConnectTimeoutSec:   99,
				HandshakeTimeoutSec: 88,
				RetryMaxCount:       77,
				RetryIntervalSec:    66,
			},
			want: &Options{
				ConnectTimeoutSec:   99,
				HandshakeTimeoutSec: 88,
				RetryMaxCount:       77,
				RetryIntervalSec:    66,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.SetDefaults()
			if !reflect.DeepEqual(tt.base, tt.want) {
				t.Errorf("Options.SetDefaults() = %v, want %v", tt.base, tt.want)
			}
		})
	}
}
