// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var devMode string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nwsctl",
	Short: "nwsctl is a CLI tool for declarative infrastructure management, automating edge cluster setup via state definitions and recipes.",
	Long: `nwsctl is a command-line interface for the Niwashi framework, designed
to automate the construction and management of edge clusters. It allows
you to define your infrastructure's desired state declaratively in YAML
files.

The typical workflow involves:
  1. Plan: Use nwsctl plan to compare your desired state (target.yaml)
     with the current state. This generates an execution plan (DAG)
     that details the necessary steps to achieve the target
     configuration.
  2. Apply: Use nwsctl apply to execute the generated plan. This runs
     the required recipes (wrapping tools like Ansible or Terraform) in
     the correct order to update your infrastructure.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.EnableCommandSorting = false

	//	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newPlanCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newExportCmd())
	rootCmd.AddCommand(newSshCmd())

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.temp.yaml)")

	rootCmd.PersistentFlags().StringVarP(&devMode, "dev-mode", "", "", "developer option (internal use only)")
	if err := rootCmd.PersistentFlags().MarkHidden("dev-mode"); err != nil {
		panic("Failed to mark --dev-mode flag as hidden: " + err.Error())
	}
}

var devOptions map[string]string

func GetDevOptions() map[string]string {
	if devOptions == nil {
		devOptions = make(map[string]string)
		if devMode != "" {
			// parse devMode string like "key1=value1,key2=value2"
			pairs := strings.Split(devMode, ",")
			for _, pair := range pairs {
				kv := strings.Split(pair, "=")
				value := ""
				if len(kv) > 1 {
					value = strings.Join(kv[1:], "=")
				}
				devOptions[kv[0]] = value
			}
		}
	}

	return devOptions
}
