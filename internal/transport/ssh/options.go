// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package ssh

import (
	"os"
	"strconv"
	"time"
)

type Options struct {
	ConnectTimeoutSec   int `json:"connectTimeoutSec,omitempty" yaml:"connectTimeoutSec,omitempty"`
	HandshakeTimeoutSec int `json:"handshakeTimeoutSec,omitempty" yaml:"handshakeTimeoutSec,omitempty"`
	RetryMaxCount       int `json:"retryMaxCount,omitempty" yaml:"retryMaxCount,omitempty"`
	RetryIntervalSec    int `json:"retryIntervalSec,omitempty" yaml:"retryIntervalSec,omitempty"`
}

var (
	defaultOptions = &Options{
		ConnectTimeoutSec:   10,
		HandshakeTimeoutSec: 10,
		RetryMaxCount:       3,
		RetryIntervalSec:    5,
	}
	envOptions *Options
)

func init() {
	envOptions = getEnvOptions()
}

func getEnvOptions() *Options {
	opts := &Options{}
	v := os.Getenv("NWS_SSH_CONNECT_TIMEOUT_SEC")
	if v != "" {
		if t, err := time.ParseDuration(v + "s"); err == nil {
			opts.ConnectTimeoutSec = int(t.Seconds())
		}
	}

	v = os.Getenv("NWS_SSH_HANDSHAKE_TIMEOUT_SEC")
	if v != "" {
		if t, err := time.ParseDuration(v + "s"); err == nil {
			opts.HandshakeTimeoutSec = int(t.Seconds())
		}
	}

	v = os.Getenv("NWS_SSH_RETRY_MAX_COUNT")
	if v != "" {
		if n, err := strconv.ParseInt(v, 10, 0); err == nil {
			opts.RetryMaxCount = int(n)
		}
	}

	v = os.Getenv("NWS_SSH_RETRY_INTERVAL_SEC")
	if v != "" {
		if t, err := time.ParseDuration(v + "s"); err == nil {
			opts.RetryIntervalSec = int(t.Seconds())
		}
	}
	return opts
}

func (o *Options) SetDefaults() {
	if o.ConnectTimeoutSec == 0 {
		o.ConnectTimeoutSec = defaultOptions.ConnectTimeoutSec
	}
	if o.HandshakeTimeoutSec == 0 {
		o.HandshakeTimeoutSec = defaultOptions.HandshakeTimeoutSec
	}
	if o.RetryMaxCount == 0 {
		o.RetryMaxCount = defaultOptions.RetryMaxCount
	}
	if o.RetryIntervalSec == 0 {
		o.RetryIntervalSec = defaultOptions.RetryIntervalSec
	}
}

func (o *Options) Clone() *Options {
	if o == nil {
		return nil
	}
	return &Options{
		ConnectTimeoutSec:   o.ConnectTimeoutSec,
		HandshakeTimeoutSec: o.HandshakeTimeoutSec,
		RetryMaxCount:       o.RetryMaxCount,
		RetryIntervalSec:    o.RetryIntervalSec,
	}
}

func (o *Options) Merge(other *Options) *Options {
	if o == nil && other == nil {
		return nil
	}
	if o == nil {
		return other.Clone()
	}
	if other == nil {
		return o.Clone()
	}
	return &Options{
		ConnectTimeoutSec: func() int {
			if other.ConnectTimeoutSec != 0 {
				return other.ConnectTimeoutSec
			}
			return o.ConnectTimeoutSec
		}(),
		HandshakeTimeoutSec: func() int {
			if other.HandshakeTimeoutSec != 0 {
				return other.HandshakeTimeoutSec
			}
			return o.HandshakeTimeoutSec
		}(),
		RetryMaxCount: func() int {
			if other.RetryMaxCount != 0 {
				return other.RetryMaxCount
			}
			return o.RetryMaxCount
		}(),
		RetryIntervalSec: func() int {
			if other.RetryIntervalSec != 0 {
				return other.RetryIntervalSec
			}
			return o.RetryIntervalSec
		}(),
	}
}

func (o *Options) ConnectTimeout() time.Duration {
	if envOptions != nil && envOptions.ConnectTimeoutSec != 0 {
		return time.Duration(envOptions.ConnectTimeoutSec) * time.Second
	}
	return time.Duration(o.ConnectTimeoutSec) * time.Second
}

func (o *Options) HandshakeTimeout() time.Duration {
	if envOptions != nil && envOptions.HandshakeTimeoutSec != 0 {
		return time.Duration(envOptions.HandshakeTimeoutSec) * time.Second
	}
	return time.Duration(o.HandshakeTimeoutSec) * time.Second
}

func (o *Options) RetryMax() int {
	if envOptions != nil && envOptions.RetryMaxCount != 0 {
		return envOptions.RetryMaxCount
	}
	return o.RetryMaxCount
}

func (o *Options) RetryInterval() time.Duration {
	if envOptions != nil && envOptions.RetryIntervalSec != 0 {
		return time.Duration(envOptions.RetryIntervalSec) * time.Second
	}
	return time.Duration(o.RetryIntervalSec) * time.Second
}
