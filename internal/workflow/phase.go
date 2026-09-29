// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import "slices"

type Phase struct {
	Priority     int
	AllowedPhase []string
}

const (
	SubjectSystemPrefix  = "#"
	PhaseBarrier         = "barrier"
	PhaseBarrierPriority = 9999

	SubjectHostPrefix = "!"
	PhaseHost         = "host"
	PhaseHostPriority = 10

	SubjectInfraPrefix = "$"
	PhaseInfra         = "infra"
	PhaseInfraPriority = 20

	PhaseNodeSync         = "node_sync"
	PhaseNodeSyncPriority = PhaseNodePriority - 5 // before node phase

	SubjectNodePrefix = "@"
	PhaseNode         = "node"
	PhaseNodePriority = 30

	PhaseClusterSync         = "cluster_sync"
	PhaseClusterSyncPriority = PhaseClusterPriority - 1 // before cluster phase

	SubjectClusterPrefix = "%"
	PhaseCluster         = "cluster"
	PhaseClusterPriority = 40
)

// lower number is higher priority
var phaseData = map[string]*Phase{}

func AddPhase(phase string, priority int, allowedPhase []string) {
	// add self to allowed phase
	allowedPhase = append(allowedPhase, phase)

	phaseData[phase] = &Phase{
		Priority:     priority,
		AllowedPhase: allowedPhase,
	}
}

// phaseA can depend on phaseB?
func CanDepends(phaseA, phaseB string) bool {

	pa, oka := phaseData[phaseA]
	pb, okb := phaseData[phaseB]
	if !oka || !okb {
		return false
	}

	if !slices.Contains(pa.AllowedPhase, phaseB) {
		return false
	}

	return pa.Priority >= pb.Priority
}

func GetOrderedPhase() []string {
	phases := make([]string, 0, len(phaseData))
	for p := range phaseData {
		phases = append(phases, p)
	}

	// sort
	for i := 0; i < len(phases); i++ {
		for j := 0; j < len(phases)-1-i; j++ {
			if phaseData[phases[j]].Priority > phaseData[phases[j+1]].Priority {
				phases[j], phases[j+1] = phases[j+1], phases[j]
			}
		}
	}

	return phases
}
