// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_barrier

import (
	"github.com/sony/niwashi/internal/workflow"
)

func init() {
	workflow.RegisterTaskGeneratorFactory(
		workflow.PhaseBarrier,
		NewBarrierTaskGenerator,
	)
}

func NewBarrierTaskGenerator(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {

	// logger.Info("Phase barrier", "phase", job.GetOptions()[0])

	return workflow.NewBasicTaskGenerator(
		ctx,
		job,
	), nil
}
