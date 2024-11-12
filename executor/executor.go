// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package executor

import (
	"context"
	"fmt"

	"vilks.io/vilks/evidence"
	"vilks.io/vilks/logger"
	"vilks.io/vilks/recipe"
)

type Executor struct {
	recipes *recipe.Recipes
	attacks []*Attack
	log     logger.Logger
	ev      evidence.Evidence

	AttackerHost string
	TeamName     string
	TeamIndex    string
}

func New(log logger.Logger, ev evidence.Evidence, recipes *recipe.Recipes) *Executor {
	return &Executor{
		log:     log,
		recipes: recipes,
		ev:      ev,
	}
}

func (e *Executor) AddAttack(host, recipe string, params map[string]string) error {
	r := e.recipes.Get(recipe)
	if r == nil {
		return fmt.Errorf("recipe '%s' not found", recipe)
	}

	e.attacks = append(e.attacks, &Attack{
		executor: e,

		Host:     host,
		Recipe:   r,
		Params:   params,
		Evidence: make(map[string]string),
	})

	return nil
}

func (e *Executor) Validate(_ context.Context) error {
	for _, a := range e.attacks {
		for _, p := range a.Recipe.Params {
			if v, ok := a.Params[p.Name]; p.Required && (!ok || len(v) == 0) {
				return fmt.Errorf("missing required parameter '%s' for recipe '%s'", p.Name, a.Recipe.Name)
			}
		}
	}

	return nil
}

func (e *Executor) Execute(ctx context.Context) error {
	for _, a := range e.attacks {
		e.log.Info(fmt.Sprintf("Executing recipe '%s' on host '%s'", e.log.Special(a.Recipe.Name), e.log.Special(a.Host)), a.Values())

		if err := a.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}
