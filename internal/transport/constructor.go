// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport

import "fmt"

var constructors = map[string]func() Transport{}

func RegisterConstructor(
	transportType string,
	constructor func() Transport,
) {
	constructors[transportType] = constructor
}

func NewTransport(transportType string) (Transport, error) {
	constructor, ok := constructors[transportType]
	if !ok {
		return nil, fmt.Errorf("unsupported transport type: %s", transportType)
	}
	return constructor(), nil
}
