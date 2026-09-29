// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package cmd

import (
	"fmt"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
	"github.com/spf13/cobra"
)

type PlanData struct {
	WorkDir   string
	Recipes   []string
	Profile   string
	Targets   []string
	Out       string
	Prune     bool
	Destroy   bool
	WithInit  bool
	LogConfig logger.Config
}

var planData PlanData

func newPlanCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Create an execution plan",
		Long: `Generates an execution plan by comparing the initial and target states.
It analyzes differences, constructs a DAG of tasks based on recipes, and
saves the plan to a JSON file for execution.`,
		SilenceUsage: true, // Don't show usage on error
		RunE: func(cmd *cobra.Command, args []string) error {

			initPlanLogger(planData.LogConfig)
			defer finalizePlanLogger()

			err := RunPlanCmd(&planData)
			if err != nil {
				logger.Error("plan command failed:", "error", err)
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&planData.WorkDir, FlagWorkDir, DefaultWorkDir, UsageWorkDir)
	cmd.Flags().StringArrayVar(&planData.Recipes, FlagRecipeDir, DefaultRecipeDir, UsageRecipeDir)
	cmd.Flags().StringVar(&planData.Profile, "profile", "", "Path to a plan profile file.")
	cmd.Flags().BoolVar(&planData.WithInit, "with-init", false, "Include the initial state in the plan.")
	cmd.Flags().StringArrayVarP(&planData.Targets, "target", "t", []string{}, "Path to one or more target state files.")
	cmd.Flags().StringVarP(&planData.Out, "out", "o", "plan.json", "Path to save the generated plan file in JSON format.")
	cmd.Flags().BoolVar(&planData.Prune, "prune", false, "Create a plan that only includes destroy operations.")
	cmd.Flags().BoolVar(&planData.Destroy, "destroy", false, "Create a plan to destroy all resources.")

	// --log-file=[path|-]
	// --log-format=[plain|json]
	cmd.Flags().StringVar(&planData.LogConfig.LogFile, "log-file", "", "Path to the log file. Use '-' for stdout.")
	cmd.Flags().StringVar(&planData.LogConfig.LogFormat, "log-format", "plain", "Log format (plain or json).")

	return cmd
}

func RunPlanCmd(arg *PlanData) error {

	logger.Progress(
		workflow.PlanStart.String(),
		"targets", arg.Targets,
		"profile", arg.Profile,
		"destroy", arg.Destroy,
		"prune", arg.Prune)

	fs := file.NewDefaultFileSystem()

	if arg.WithInit {
		if err := RunInitCmd(&InitData{
			WorkDir:    arg.WorkDir,
			filesystem: fs,
		}); err != nil {
			return err
		}
	}

	opts := []func(*planner.Config){
		planner.WithWorkDir(arg.WorkDir),
		planner.WithRecipeDirs(arg.Recipes...),
		planner.WithProfile(arg.Profile),
	}

	if arg.Destroy && arg.Prune {
		return fmt.Errorf("cannot specify both --destroy and --prune options")
	}

	if arg.Destroy {
		if len(arg.Targets) > 0 {
			logger.Error("target files are ignored when --destroy option is specified")
		}
		opts = append(opts, planner.WithDestroy())
	} else {
		opts = append(opts, planner.WithTargetFiles(arg.Targets...))
	}

	// Disable construct phase in prune option
	if arg.Prune {
		opts = append(opts, planner.WithPrune())
	}

	config := planner.NewConfig(opts...)
	p, err := planner.NewPlanner(config)
	if err != nil {
		return err
	}

	// Make plan
	planData, err := p.MakePlan()
	if err != nil {
		logger.Error("failed to make plan:", "error", err)
		return err
	}

	// Save plan to file
	err = plan.Write(planData, arg.Out, fs)
	if err != nil {
		logger.Error("failed to save plan:", "error", err)
		return err
	}
	planData.FilePath = arg.Out

	logger.Progress(
		workflow.PlanComplete.String(),
		"out", arg.Out,
		"error", err)

	ShowPlan(planData)

	return nil
}

func ShowPlan(p *plan.Plan) {
	fmt.Printf("------\n")
	fmt.Printf("Plan Summary:\n")
	// fmt.Printf("- Working Directory: %s\n", func() string { d, _ := os.Getwd(); return d }())
	// fmt.Printf("- Recipe Directories: %v\n", p.RecipeDirs)
	fmt.Printf("- Mode: %s\n", p.Mode)

	var jobs []*plan.Job
	if p.Mode == plan.ModeDefault {
		jobs = p.ExecutionPlan.ConstructGraph.Jobs
	} else {
		jobs = p.ExecutionPlan.DestructGraph.Jobs
	}
	fmt.Printf("- Total Jobs: %d\n", len(jobs))
	fmt.Printf("- Plan file saved at: %s\n", p.FilePath)
	fmt.Printf("- Initial state hash: %s\n", p.States.Initial.Hash)

	fmt.Printf("\nJobs:\n")
	for i, job := range jobs {
		fmt.Printf("%2d. %s\n", i+1, job.Id)
	}

	if len(p.PendingUpdates) > 0 {
		fmt.Printf("\nPending Updates (recipe has no update tasks):\n")
		for _, recipe := range p.PendingUpdates {
			fmt.Printf("- %v\n", recipe)
		}
	}

	if len(p.Runtime.Bindings) > 0 {
		fmt.Printf("\nRecipe Bindings:\n")
		for key, alias := range p.Runtime.Bindings {
			fmt.Printf("- %s -> %s\n", key, alias)
		}
	}

	if len(p.Runtime.ToolAlias) > 0 {
		fmt.Printf("\nTool Aliases:\n")
		for key, alias := range p.Runtime.ToolAlias {
			fmt.Printf("- %s -> %s\n", key, alias)
		}
	}
}
