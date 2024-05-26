// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"errors"
	"fmt"
	"os"

	"vilks.io/vilks/scenario"

	"github.com/spf13/cobra"
)

var (
	evidencePath string

	attackerIP string
	teamName   string
	hostName   string
	attackName string
)

func runExec(cmd *cobra.Command, _ []string) error {
	if attackerIP == "" {
		ip, err := getHostIP()
		if err != nil {
			return fmt.Errorf("failed to get host IP: %s", err.Error())
		}

		attackerIP = ip
	}

	if scenarioPath == "" {
		return errors.New("scenario file is required")
	}

	if recipesDir == "" {
		return errors.New("recipes directory is required")
	}

	if evidencePath == "" {
		return errors.New("evidence directory is required")
	}

	log.Info("Loading scenario...")

	scene, err := scenario.New(cmd.Context(), log, evidencePath, scenarioPath, recipesDir)
	if err != nil {
		log.Error("Failed to load scenario: " + err.Error())
		os.Exit(1)
	}

	scene.SetAttackerHost(attackerIP)

	for _, team := range scene.Teams() {
		if teamName != "" && team.Name != teamName {
			continue
		}

		for _, host := range scene.Hosts() {
			if hostName != "" && host != hostName {
				continue
			}

			for _, attack := range scene.Attacks(host) {
				if attackName != "" && attack != attackName {
					continue
				}

				log.Info("Team: " + log.Special(team.Name))
				log.Info("Host: " + log.Special(host))
				log.Info("Starting attack " + log.Special(attack))

				if err := scene.Execute(cmd.Context(), team.Name, host, attack); err != nil {
					log.Error("Attack failed: " + err.Error())
				} else {
					log.Info("Attack completed")
				}

				log.Info("Scenario completed")
			}
		}
	}

	return nil
}

func init() {
	initRootCmd()

	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute",
		Long:  `Execute scenario with given parameters.`,
		RunE:  runExec,
	}

	cmd.Flags().StringVarP(&scenarioPath, "scenario", "s", scenarioPath, "Path to scenario file")
	_ = cmd.MarkFlagRequired("scenario")
	cmd.Flags().StringVarP(&recipesDir, "recipes", "r", recipesDir, "Path to recipes directory")
	_ = cmd.MarkFlagRequired("recipes")
	cmd.Flags().StringVarP(&evidencePath, "evidence", "e", evidencePath, "Path to evidence directory")
	_ = cmd.MarkFlagRequired("evidence")

	cmd.Flags().StringVarP(&attackerIP, "attacker", "a", attackerIP, "Attacker IP address")
	cmd.Flags().StringVar(&teamName, "team", teamName, "Team name")
	cmd.Flags().StringVar(&hostName, "host", hostName, "Host name")
	cmd.Flags().StringVar(&attackName, "attack", attackName, "Attack name")

	RootCmd.AddCommand(cmd)
}
