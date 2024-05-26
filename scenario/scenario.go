// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package scenario

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"vilks.io/vilks/evidence"
	"vilks.io/vilks/executor"
	"vilks.io/vilks/logger"
	"vilks.io/vilks/recipe"
)

type Scene struct {
	attackerHost string
	scenario     *Scenario
	recipes      *recipe.Recipes
	log          logger.Logger
	evmgr        *evidence.Manager
}

func New(ctx context.Context, log logger.Logger, evidencePath, scenarioPath, recipesDir string) (*Scene, error) {
	scenario, err := LoadFile(ctx, scenarioPath)
	if err != nil {
		return nil, err
	}

	recipes := recipe.New()

	err = filepath.Walk(recipesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		return recipes.AddFromPath(path)
	})
	if err != nil {
		return nil, err
	}

	return &Scene{
		scenario: scenario,
		recipes:  recipes,
		log:      log,
		evmgr:    evidence.New(evidencePath),
	}, nil
}

func (s *Scene) Validate(_ context.Context) error {
	for _, h := range s.scenario.Hosts {
		for _, a := range h.Attacks {
			if s.recipes.Get(a.Recipe) == nil {
				return fmt.Errorf("recipe '%s' not found", a.Recipe)
			}
		}
	}

	return nil
}

func (s *Scene) SetAttackerHost(host string) {
	s.attackerHost = host
}

func (s *Scene) Teams() []Team {
	return s.scenario.Teams
}

func (s *Scene) Hosts() []string {
	hosts := make([]string, 0, len(s.scenario.Hosts))

	for _, h := range s.scenario.Hosts {
		hosts = append(hosts, h.Name)
	}

	return hosts
}

func (s *Scene) Attacks(host string) []string {
	for _, h := range s.scenario.Hosts {
		if h.Name == host {
			attacks := make([]string, 0, len(h.Attacks))

			for _, a := range h.Attacks {
				attacks = append(attacks, a.Name)
			}

			return attacks
		}
	}

	return nil
}

func (s *Scene) Execute(ctx context.Context, teamName, hostName, attackName string) error {
	var team *Team

	for i, t := range s.scenario.Teams {
		if t.Name == teamName {
			team = &s.scenario.Teams[i]

			break
		}
	}

	if team == nil {
		return fmt.Errorf("team '%s' not found", teamName)
	}

	var host *Host

	for i, h := range s.scenario.Hosts {
		if h.Name == hostName {
			host = &s.scenario.Hosts[i]

			break
		}
	}

	if host == nil {
		return fmt.Errorf("host '%s' not found", hostName)
	}

	var attack *Attack

	for i, a := range host.Attacks {
		if a.Name == attackName {
			attack = &host.Attacks[i]

			break
		}
	}

	if attack == nil {
		return fmt.Errorf("host '%s' attack '%s' not found", hostName, attackName)
	}

	ex := executor.New(s.log, s.evmgr.Attack(team.Name, host.Name), s.recipes)
	ex.AttackerHost = s.attackerHost
	ex.TeamName = team.Name
	ex.TeamIndex = strconv.FormatInt(int64(team.Index), 10)

	target := strings.ReplaceAll(host.Target, "{x}", strconv.FormatInt(int64(team.Index), 10))
	params := make(map[string]string, len(attack.Params))

	for _, prm := range attack.Params {
		params[prm.Name] = strings.ReplaceAll(prm.Value, "{x}", strconv.FormatInt(int64(team.Index), 10))
	}

	if err := ex.AddAttack(target, attack.Recipe, params); err != nil {
		return err
	}

	if err := ex.Validate(ctx); err != nil {
		return err
	}

	return ex.Execute(ctx)
}
