// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node_sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/applier"
)

func buildNodeBindTask(ctx workflow.ApplyContext, job *workflow.Job) (workflow.Task, error) {
	switch ctx.Operation() {
	case recipe.OperationConstruct:
		return NewNodeBindTask(ctx, job)
	case recipe.OperationDestruct:
		return NewNodeUnbindTask(ctx, job)
	}
	return nil, fmt.Errorf("unsupported operation for NodeSyncTask: %s", ctx.Operation())
}

type SelectorInfo struct {
	NodeName string
	Selector *instance.Selector
}

type NodeBindTask struct {
	applier.SystemTask
	ctx   workflow.ApplyContext
	Nodes []*SelectorInfo
}

func NewNodeBindTask(ctx workflow.ApplyContext, job *workflow.Job) (*NodeBindTask, error) {

	nodes, exist := job.GetFromData("nodes")
	if !exist {
		return nil, fmt.Errorf("invalid data in NodeBindTask")
	}

	// convert nodes to map[string]*instance.Selector
	nodeMap, ok := nodes.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid type for nodes, expected types.Dict")
	}
	syncNodes := []*SelectorInfo{}
	for nodeName, raw := range nodeMap {
		var selector *instance.Selector
		if raw != nil {
			jsonData, err := json.Marshal(raw)
			if err != nil {
				return nil, err
			}
			var s instance.Selector
			err = json.Unmarshal(jsonData, &s)
			if err != nil {
				return nil, err
			}
			selector = &s
		} else {
			selector = &instance.Selector{}
		}
		syncNodes = append(syncNodes, &SelectorInfo{
			NodeName: nodeName,
			Selector: selector,
		})
	}

	return &NodeBindTask{
		ctx:        ctx,
		SystemTask: applier.NewSystemTask("bind nodes", job),
		Nodes:      syncNodes,
	}, nil
}

func (ct *NodeBindTask) Run(_ context.Context) error {

	ct.Nodes = sort(ct.Nodes)

	return nil
}

func sort(originalNodes []*SelectorInfo) []*SelectorInfo {
	sorted := []*SelectorInfo{}
	nodes := slices.Clone(originalNodes)

	selectInfo := func(comp func(i *SelectorInfo) bool) {
		remains := []*SelectorInfo{}
		for len(nodes) > 0 {
			sel := nodes[0]
			if comp(sel) {
				sorted = append(sorted, sel)
			} else {
				remains = append(remains, sel)
			}
			nodes = nodes[1:]
		}
		nodes = remains
	}

	// 1. Bind nodes who have specific generator and instance
	selectInfo(func(i *SelectorInfo) bool {
		return i.Selector.Generator != "" && i.Selector.Instance != ""
	})

	// 2. Bind nodes who have specific generator
	selectInfo(func(i *SelectorInfo) bool {
		return i.Selector.Generator != "" && i.Selector.Instance == ""
	})

	// 3. Bind nodes who do not have specific generator
	selectInfo(func(i *SelectorInfo) bool {
		return true
	})

	return sorted
}

func (ct *NodeBindTask) AfterRun(runErr error) error {
	if runErr != nil {
		return nil
	}

	// init
	patches := []patch.Patch{}
	s := ct.ctx.State()
	binder := instance.NewBinder(s.Infrastructure.Generators)

	for _, i := range ct.Nodes {

		if bound, err := ct.BoundAlready(i.NodeName); bound {
			if err != nil {
				// bound but error occurs, so return error
				return errors.Join(err, fmt.Errorf("failed to check if node %q is already bound", i.NodeName))
			}
			logger.Info("node already bound, skipping", "node", i.NodeName)
			continue
		}

		genName, instName, err := binder.Bind(i.Selector)
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to find instance for node %q", i.NodeName))
		}

		nodePath := "/inventory/nodes/" + i.NodeName

		// if node does not exist, create empty node
		if s.GetNode(i.NodeName) == nil {
			// add empty node
			patches = append(
				patches, patch.NewAddPatch(
					nodePath,
					node.NewNode(),
				))
		}

		patches = append(
			patches,
			patch.NewAddPatch(
				"/infrastructure/generators/"+genName+"/instances/"+instName+"/nodeRef",
				i.NodeName,
			),
			patch.NewAddPatch(
				nodePath+"/instanceRef",
				genName+":"+instName,
			),
		)

		tpPatches, err := updateTransport(ct.ctx, s.GetGenerator(genName).GetInstance(instName))
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to update transport for node %q", i.NodeName))
		}
		patches = append(
			patches,
			tpPatches...,
		)
	}

	return ct.ctx.ApplyPatches(patches, nil, nil)
}

func (ct *NodeBindTask) BoundAlready(nodeName string) (bool, error) {

	inst := ct.ctx.State().GetBoundInstance(nodeName)
	if inst == nil {
		return false, nil
	}

	if inst.NodeRef == nodeName {
		return true, nil
	}

	return true, fmt.Errorf("node %q is already bound to instance %s:%s, but nodeRef is %q", nodeName, inst.GeneratorName, inst.InstanceName, inst.NodeRef)
}

type NodeUnbindTask struct {
	applier.SystemTask
	ctx   workflow.ApplyContext
	Nodes []string
}

func NewNodeUnbindTask(ctx workflow.ApplyContext, job *workflow.Job) (*NodeUnbindTask, error) {
	nodes, exist := job.GetFromData("nodes")
	if !exist {
		return nil, fmt.Errorf("invalid data in NodeUnbindTask")
	}
	nodeMap, ok := nodes.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid type for nodes, expected types.Dict")
	}

	return &NodeUnbindTask{
		ctx:        ctx,
		SystemTask: applier.NewSystemTask("unbind nodes", job),
		Nodes:      slices.Collect(maps.Keys(nodeMap)),
	}, nil
}

func (dt *NodeUnbindTask) Run(_ context.Context) error {
	// do nothing
	return nil
}

func (dt *NodeUnbindTask) AfterRun(runErr error) error {
	if runErr != nil {
		return nil
	}

	patches := []patch.Patch{}
	for _, nodeName := range dt.Nodes {

		// skip if node not found
		if dt.ctx.State().GetNode(nodeName) == nil {
			logger.Warn("node not found, skipping", "node", nodeName)
			continue
		}

		nodePatches, err := dt.CreateUnbindPatch(nodeName)
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to create unbind patch for node %q", nodeName))
		}
		if len(nodePatches) != 0 {
			patches = append(patches, nodePatches...)
		}

		// remove node
		patches = append(patches, patch.NewRemovePatch(
			fmt.Sprintf("/inventory/nodes/%s", nodeName),
		))

		// Remove entry from known_hosts
		inst := dt.ctx.State().GetBoundInstance(nodeName)
		if inst == nil {
			return fmt.Errorf("failed to get instance for node %q", nodeName)
		}

		if inst.NodeRef != nodeName {
			return fmt.Errorf("instance %s:%s is not bound to node %s", inst.GeneratorName, inst.InstanceName, nodeName)
		}

		tpPatches, err := updateTransport(dt.ctx, inst)
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to update transport for node %q", nodeName))
		}
		if len(tpPatches) != 0 {
			patches = append(patches, tpPatches...)
		}
	}

	return dt.ctx.ApplyPatches(patches, nil, nil)
}

func (dt *NodeUnbindTask) CreateUnbindPatch(nodeName string) ([]patch.Patch, error) {
	patches := []patch.Patch{}

	inst := dt.ctx.State().GetBoundInstance(nodeName)
	if inst == nil {
		return nil, fmt.Errorf("instance not found for node %q", nodeName)
	}

	if inst.NodeRef != nodeName {
		return nil, fmt.Errorf("instance %s:%s is not bound to node %s", inst.GeneratorName, inst.InstanceName, nodeName)
	}

	patches = append(patches, patch.NewRemovePatch(
		fmt.Sprintf("/inventory/nodes/%s/instanceRef", nodeName),
	))

	patches = append(
		patches,
		patch.NewRemovePatch(
			fmt.Sprintf("/infrastructure/generators/%s/instances/%s/nodeRef", inst.GeneratorName, inst.InstanceName),
		))

	return patches, nil
}

func updateTransport(ctx workflow.ApplyContext, inst *instance.InstanceAccessor) ([]patch.Patch, error) {

	// todo: implement "enroll" sub command to update transport info
	return nil, nil
}
