// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package winrm

const defaultConnectTimeoutSec = 30

type Options struct {
	// AllowInsecureHTTP opts into plain HTTP (port 5985) instead of the
	// default HTTPS (port 5986). Zero value (false) keeps HTTPS required,
	// so a recipe author has to explicitly ask for the insecure transport.
	AllowInsecureHTTP bool `json:"allowInsecureHttp,omitempty" yaml:"allowInsecureHttp,omitempty"`
	ConnectTimeoutSec int  `json:"connectTimeoutSec,omitempty" yaml:"connectTimeoutSec,omitempty"`
}

func (o *Options) UseHTTPS() bool {
	return !o.AllowInsecureHTTP
}

func (o *Options) SetDefaults() {
	if o.ConnectTimeoutSec == 0 {
		o.ConnectTimeoutSec = defaultConnectTimeoutSec
	}
}

func (o *Options) Clone() *Options {
	if o == nil {
		return nil
	}
	return &Options{
		AllowInsecureHTTP: o.AllowInsecureHTTP,
		ConnectTimeoutSec: o.ConnectTimeoutSec,
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
		AllowInsecureHTTP: o.AllowInsecureHTTP || other.AllowInsecureHTTP,
		ConnectTimeoutSec: func() int {
			if other.ConnectTimeoutSec != 0 {
				return other.ConnectTimeoutSec
			}
			return o.ConnectTimeoutSec
		}(),
	}
}
