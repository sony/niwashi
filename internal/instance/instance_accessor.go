// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import "github.com/sony/niwashi/internal/transport"

type InstanceAccessor struct {
	*Instance
	GeneratorName, InstanceName string
}

func (a *InstanceAccessor) GetFqid() string {
	return a.GeneratorName + ":" + a.InstanceName
}

func NewInstanceAccessor(generatorName, instanceName string, i *Instance) *InstanceAccessor {
	if i == nil {
		return nil
	}
	return &InstanceAccessor{
		GeneratorName: generatorName,
		InstanceName:  instanceName,
		Instance:      i,
	}
}

func (a *InstanceAccessor) GetConnection() *transport.Connection {
	if a == nil || a.Instance == nil {
		return nil
	}
	return a.Connection
}
