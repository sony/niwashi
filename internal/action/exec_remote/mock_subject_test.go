// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/transport"
)

type subject struct {
	scope string
	id    string
	conn  *transport.Connection
	op    transport.Operator
	os    string
}

func NewSubject(scope, id string, opts ...func(*subject)) *subject {
	s := &subject{
		scope: scope,
		id:    id,
		os:    platform.OsUnknown,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithConnection(conn *transport.Connection) func(*subject) {
	return func(s *subject) {
		s.conn = conn
	}
}

func WithOperator(op transport.Operator) func(*subject) {
	return func(s *subject) {
		s.op = op
	}
}

func WithOs(os string) func(*subject) {
	return func(s *subject) {
		s.os = os
	}
}

func (s *subject) GetScope() string {
	return s.scope
}

func (s *subject) GetId() string {
	return s.id
}

func (s *subject) GetOs() string {
	return s.os
}

func (s *subject) GetTransport() (transport.Transport, error) {
	return s.conn.GetTransport("")
}

func (s *subject) GetOperator(t transport.Transport) (transport.Operator, error) {
	if s.op != nil {
		return s.op, nil
	}
	return transport.NewOperator(t.GetType())
}
