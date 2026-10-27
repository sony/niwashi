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
	"github.com/sony/niwashi/internal/state"
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
	if err := ValidateState(ctx.RecipeFinder(), ctx.TargetState()); err != nil {
		return nil, err
	}
	if err := ValidateState(ctx.RecipeFinder(), ctx.State()); err != nil {
		return nil, err
	}
	if err := ValidatePlan(ctx.RecipeFinder(), ctx.Plan()); err != nil {
		return nil, err
	}

	return &Applier{
		applyContext: ctx,
	}, nil
}

func modeName(mode plan.Mode) string {
	switch mode {
	case plan.ModeDefault:
		return "construct"
	case plan.ModeDestroy:
		return "destroy"
	case plan.ModePrune:
		return "prune"
	default:
		return "unknown"
	}
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

	rootCtx := a.applyContext

	logger.Info("Applying " + modeName(rootCtx.Plan().Mode) + " plan...")

	// for each stage
	stages, err := a.applyContext.Plan().StageOrder()
	if err != nil {
		return err
	}

	for _, stage := range stages {
		ctx, err := rootCtx.WithStage(stage)
		if err != nil {
			return err
		}

		if err := runStage(ctx, stage); err != nil {
			return err
		}

		if err := syncExportData(ctx); err != nil {
			return err
		}
	}
	return nil
}

func runStage(ctx *applyContext, stage string) error {
	a := &Applier{
		applyContext: ctx,
	}

	if ctx.workflowGraph == nil {
		logger.Info("No workflow:", "stage", stage)
		return nil
	}

	jobQueue := workflow.NewSingleJobQueue(ctx.workflowGraph)

	engine := workflow.NewEngine(
		workflow.WithMaxConcurrency(a.applyContext.applyCfg.Concurrency),
		workflow.WithScheduler(workflow.NewTaskScheduler(ctx, jobQueue)),
		workflow.WithListener(a),
	)

	return engine.Run(context.Background())
}

func syncExportData(ctx *applyContext) error {

	stage := ctx.StageName()
	if stage == state.InitStageName || stage == state.MainStageName {
		// _init and main stages have no exports
		return nil
	}

	currentStage, _ := ctx.state.FindStage(stage)
	targetStage, existTargetStage := ctx.targetState.FindStage(stage)

	var exports state.Exports

	switch ctx.Plan().Mode {
	case plan.ModeDefault:
		// In construct mode, merge exports from the target stage into the current stage

		if !existTargetStage {
			// do nothing
			return nil
		}
		exports = currentStage.ExportMerge(targetStage)

	case plan.ModePrune:
		// In prune mode, restrict exports to only those present in the target stage

		if currentStage == nil || currentStage.Export == nil {
			return nil
		}

		exports = currentStage.ExportRestrictTo(targetStage)

	case plan.ModeDestroy:
		// In destroy mode, clear all exports

		if currentStage == nil || currentStage.Export == nil {
			return nil
		}

		exports = nil
	}

	if err := ctx.SetExports(exports); err != nil {
		return errors.Join(err, fmt.Errorf("failed to set exports"))
	}

	return ctx.SaveState()
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
