// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"errors"
	"os"

	"vilks.io/vilks/scenario"

	"github.com/spf13/cobra"
)

var (
	scenarioPath string
	recipesDir   string
)

func runValidate(cmd *cobra.Command, _ []string) error {
	if scenarioPath == "" {
		return errors.New("scenario are required")
	}

	if recipesDir == "" {
		return errors.New("recipes are required")
	}

	scene, err := scenario.New(cmd.Context(), log, "", scenarioPath, recipesDir)
	if err != nil {
		log.Error("Failed to load scenario: " + err.Error())
		os.Exit(1)
	}

	if err := scene.Validate(cmd.Context()); err != nil {
		log.Error("Scenario contains errors: " + err.Error())
		os.Exit(1)
	}

	log.Info("Scenario is valid")

	return nil
}

func init() {
	initRootCmd()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate",
		Long:  `Validate scenario.`,
		RunE:  runValidate,
	}

	cmd.Flags().StringVarP(&scenarioPath, "scenario", "s", scenarioPath, "Path to scenario file")
	_ = cmd.MarkFlagRequired("scenario")
	cmd.Flags().StringVarP(&recipesDir, "recipes", "r", recipesDir, "Path to recipes directory")
	_ = cmd.MarkFlagRequired("recipes")

	RootCmd.AddCommand(cmd)
}
