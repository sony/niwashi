// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

type GeneratorAccessor struct {
	*Generator
	Name string
}

func NewGeneratorAccessor(name string, g *Generator) *GeneratorAccessor {
	if g == nil {
		return nil
	}
	return &GeneratorAccessor{
		Name:      name,
		Generator: g,
	}
}

func (a *GeneratorAccessor) GetInstance(instanceId string) *InstanceAccessor {
	if a == nil || a.Generator == nil {
		return nil
	}
	return NewInstanceAccessor(a.Name, instanceId, a.Instances[instanceId])
}
