// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import "context"

type Task interface {
	GetName() string
	GetSubject() string
	GetAffectSubjects() []string
	GetJob() *Job

	Run(ctx context.Context) error
	AfterRun(runError error) error
}
