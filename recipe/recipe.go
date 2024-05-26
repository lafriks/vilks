// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package recipe

import (
	"github.com/goccy/go-yaml"
)

// Recipe is a recipe for a container.
type Recipe struct {
	// Name is the name of the recipe.
	Name string `json:"name"`
	// Params is the list of input parameters for the recipe.
	Params Params `json:"params"`
	// Workspace is the list of workspace items for the recipe.
	Workspace []*WorkspaceItem `json:"workspace"`
	// Services is the list of services to run in the recipe.
	Services []*Service `json:"services"`
	// Steps is the list of steps to execute in the recipe.
	Steps []*Step `json:"steps"`
}

// Params is a map of input parameters for a recipe.
type Params map[string]*Param

func (p *Params) UnmarshalYAML(data []byte) error {
	// Unmarshal the JSON array data into a map.
	var prms []*Param
	if err := yaml.Unmarshal(data, &prms); err != nil {
		return err
	}

	*p = make(Params, len(prms))

	for _, prm := range prms {
		(*p)[prm.Name] = prm
	}

	return nil
}

// Param is an input parameter for a recipe.
type Param struct {
	// Name is the name of the parameter.
	Name string `json:"name"`
	// Description is the description of the parameter.
	Description string `json:"description"`
	// Type is the type of the parameter.
	Type string `json:"type"`
	// Required is a flag indicating if the parameter is required.
	Required bool `json:"required"`
	// Default is the default value of the parameter.
	Default string `json:"default"`
}

func Load(data []byte) (*Recipe, error) {
	var r Recipe
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type WorkspaceItem struct {
	// Source is the source path of the item.
	Source string `json:"source"`
	// Target is the target path of the item.
	Target string `json:"target"`
}
