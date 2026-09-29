// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node_sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/applier"
)

type NodeProbeTask struct {
	applier.SystemTask
	ctx        NodeProbeTaskContext
	instance   *instance.InstanceAccessor
	systemInfo *platform.SystemInfo
}

type NodeProbeTaskContext interface {
	GetJob() *workflow.Job
	GetDryRunMode() string
	GetFileSystem() file.FileSystem
	GetStateManager() workflow.StateManager
	GetInstanceInfo() (string, string)
}

func probable(c *transport.Connection) bool {
	if c == nil || len(c.Data) == 0 {
		return false
	}
	return true
}

func GetUnprobedInstances(s *state.Accessor) map[string][]string {

	notProbed := func(si *platform.SystemInfo) bool {
		return si == nil || si.Os == "" || si.Arch == ""
	}

	var unprobed = make(map[string][]string)
	for generatorName, gen := range s.Infrastructure.Generators {
		var instances []string
		for instanceName, inst := range gen.Instances {
			if notProbed(inst.SystemInfo) {
				if !probable(inst.Connection) {
					logger.Warn("instance %s:%s has no connection data, skipping probe", generatorName, instanceName)
				} else {
					// probe
					instances = append(instances, instanceName)
				}
			}
		}
		unprobed[generatorName] = instances
	}

	return unprobed
}

func instanceFqidFrom(ctx NodeProbeTaskContext) string {
	g, i := ctx.GetInstanceInfo()
	return fmt.Sprintf("%s:%s", g, i)
}

func NewNodeProbeTask(ctx NodeProbeTaskContext) (*NodeProbeTask, error) {

	generatorName, instanceName := ctx.GetInstanceInfo()
	instance := ctx.GetStateManager().State().GetGenerator(generatorName).GetInstance(instanceName)
	if instance == nil {
		return nil, fmt.Errorf("instance not found for %s", instanceFqidFrom(ctx))
	}

	fqid := instance.GetFqid()
	affectSubjects := []string{}
	if instance.NodeRef != "" {
		// instance affects node
		affectSubjects = append(
			affectSubjects, applier.MakeSubjectFqid(workflow.PhaseNode, instance.NodeRef))
	}

	t := &NodeProbeTask{
		SystemTask: applier.NewSystemTask(
			"probe:"+fqid,
			ctx.GetJob(),
			applier.WithSubject(applier.MakeSubjectFqid(workflow.PhaseInfra, fqid), affectSubjects),
			applier.WithDryRunMode(ctx.GetDryRunMode()),
			applier.WithFileSystem(ctx.GetFileSystem()),
		),
		ctx:      ctx,
		instance: instance,
	}

	return t, nil
}

func (t *NodeProbeTask) Run(ctx context.Context) error {

	id := instanceFqidFrom(t.ctx)
	conn := t.instance.GetConnection()
	if !probable(conn) {
		return fmt.Errorf("no connection data for %s", id)
	}

	var err error

	for _, tp := range conn.Data {
		prober, e := transport.NewProber(&proberCtx{
			NodeProbeTask: t,
			tp:            tp,
		})
		if e != nil {
			err = errors.Join(err, fmt.Errorf("failed to create prober for %s: %w", id, e))
			continue
		}

		if prober == nil {
			// no prober available for this transport, skip
			continue
		}

		info, e := prober.GetSystemInfo(ctx)
		if e != nil {
			err = errors.Join(err, fmt.Errorf("failed to probe %s: %w", id, e))
			continue
		}

		if info == nil {
			err = errors.Join(err, fmt.Errorf("prober returned nil system info for %s", id))
			continue
		}

		t.systemInfo = &platform.SystemInfo{
			Os:   platform.NormalizeOs(info.Os),
			Arch: platform.NormalizeArch(info.Arch),
			Data: info.Data,
		}
		return nil
	}

	return err
}

func (t *NodeProbeTask) AfterRun(runError error) error {

	if runError != nil {
		return nil
	}

	if t.systemInfo == nil {
		return fmt.Errorf("system info is nil after probing instance: %s", instanceFqidFrom(t.ctx))
	}

	path := fmt.Sprintf("/infrastructure/generators/%s/instances/%s/system",
		t.instance.GeneratorName, t.instance.InstanceName)

	// apply patch to update system info
	return t.ctx.GetStateManager().ApplyPatches([]patch.Patch{
		patch.NewAddPatch(
			path,
			t.systemInfo,
		),
	}, nil, nil)
}

type proberCtx struct {
	*NodeProbeTask
	tp transport.Transport
}

func (t *proberCtx) GetTransport() transport.Transport {
	return t.tp
}
