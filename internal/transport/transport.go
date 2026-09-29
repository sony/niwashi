// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import "github.com/sony/niwashi/internal/patch"

// Transport represents the transport information for connecting to a node.
type Transport interface {
	patch.Patchable

	GetType() string
	SetDefaults()
	Validate() error
	Clone() Transport
	Merge(other Transport) (Transport, error)

	Resolve(render func(string) (string, error)) error
}
