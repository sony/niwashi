// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workspace"
)

type Applier struct {
	applyContext *applyContext
}

func NewApplier(cfg *Config) (*Applier, error) {

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ctx, err := NewApplyContext(cfg)
	if err != nil {
		return nil, err
	}

	// Validation
	if err := ValidateState(ctx.RecipeFinder(), ctx.TargetState().State); err != nil {
		return nil, err
	}
	if err := ValidateState(ctx.RecipeFinder(), ctx.State().State); err != nil {
		return nil, err
	}
	if err := ValidatePlan(ctx.RecipeFinder(), ctx.Plan()); err != nil {
		return nil, err
	}

	return &Applier{
		applyContext: ctx,
	}, nil
}

func (a *Applier) Apply() error {

	// Prepare all data
	err := setupData(
		a.applyContext.runsWs,
		a.applyContext.Plan().FilePath,
		a.applyContext.FileSystem())
	if err != nil {
		return err
	}

	ctx := a.applyContext

	var op string
	var wf *plan.Workflow
	switch ctx.Plan().Mode {
	case plan.ModeDefault:
		logger.Info("Applying construct plan...")
		op = recipe.OperationConstruct
		wf = ctx.Plan().ExecutionPlan.ConstructGraph
	case plan.ModeDestroy:
		logger.Info("Applying destruct plan...")
		op = recipe.OperationDestruct
		wf = ctx.Plan().ExecutionPlan.DestructGraph
	case plan.ModePrune:
		logger.Info("Applying prune plan...")
		op = recipe.OperationDestruct
		wf = ctx.Plan().ExecutionPlan.DestructGraph
	default:
		return nil
	}

	ctx = ctx.WithOperation(op)

	jobQueue := workflow.NewSingleJobQueue(wf)

	engine := workflow.NewEngine(
		workflow.WithMaxConcurrency(a.applyContext.applyCfg.Concurrency),
		workflow.WithScheduler(workflow.NewTaskScheduler(ctx, jobQueue)),
		workflow.WithListener(a),
	)

	return engine.Run(context.Background())
}

func (a *Applier) OnTaskComplete(caller *workflow.Engine, res workflow.TaskResult) {
	err := a.applyContext.TaskFinished(res)
	if err != nil {
		// Handle the error if needed
		caller.RaiseError(err)
	}
}

func setupData(
	runsWs *workspace.RunWorkspace,
	srcPlanFile string,
	fs file.FileSystem) (err error) {

	err = file.MakeDirsAll(
		fs,
		runsWs.GetStateRootPath(),
		runsWs.GetStoreRootPath(),
		runsWs.GetRunsRootPath(),
		runsWs.GetEphemeralRootPath(),
	)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to create system directories"))
	}

	// Copy plan file
	planFilePath := runsWs.GetPlanFilePath()
	dest, err := fs.OpenFile(planFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to open plan file for write: %q", planFilePath))
	}
	defer func() {
		err = errors.Join(err, dest.Close())
	}()
	src, err := fs.Open(srcPlanFile)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to open plan file: %q", srcPlanFile))
	}
	defer func() { _ = src.Close() }()
	_, err = io.Copy(dest, src)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to copy plan file to: %q", planFilePath))
	}

	// Write latest file
	latestFilePath := runsWs.GetLatestFilePath()
	writer, err := fs.OpenFile(latestFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to open latest run ID file for write: %q", latestFilePath))
	}
	defer func() {
		err = errors.Join(err, writer.Close())
	}()
	_, err = writer.Write([]byte(runsWs.RunsId))
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to write latest run ID: %q", runsWs.RunsId))
	}

	return nil
}
