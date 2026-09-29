// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"

	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/transport"
)

type Subject struct {
	subTask *FilteredTask
	node    *runtime.Node
}

func NewSubject(s *runtime.State, subTask *FilteredTask) *Subject {

	node := getRuntimeNode(subTask.Scope, s, &subTask.SubjectId)

	return &Subject{
		subTask: subTask,
		node:    node,
	}
}

func (s *Subject) GetScope() string {
	return s.subTask.Scope
}

func (s *Subject) GetId() string {
	return s.subTask.Name
}

func (s *Subject) GetOs() string {
	if s.node == nil {
		return platform.OsUnknown
	}
	return s.node.Os
}

func (s *Subject) GetTransport() (transport.Transport, error) {
	if s.node == nil {
		return nil, fmt.Errorf("node not found for subject: scope=%s, id=%s", s.GetScope(), s.GetId())
	}

	// todo: return specific type from recipe if exists
	return s.node.Connection.GetTransport("")
}

func (*Subject) GetOperator(t transport.Transport) (transport.Operator, error) {
	return transport.NewOperator(t.GetType())
}
