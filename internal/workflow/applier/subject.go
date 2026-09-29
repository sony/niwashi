// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import "github.com/sony/niwashi/internal/workflow"

var prefixTable = map[string]string{
	workflow.PhaseHost:    workflow.SubjectHostPrefix,
	workflow.PhaseInfra:   workflow.SubjectInfraPrefix,
	workflow.PhaseNode:    workflow.SubjectNodePrefix,
	workflow.PhaseCluster: workflow.SubjectClusterPrefix,
}

func MakeSubjectFqid(scope, id string) string {
	if prefix, ok := prefixTable[scope]; ok {
		return prefix + id
	}

	return scope + ":" + id
}
