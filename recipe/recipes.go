// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package recipe

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Recipes struct {
	recipes map[string]*Recipe
}

// New creates a collection of recipes.
func New() *Recipes {
	return &Recipes{
		recipes: make(map[string]*Recipe),
	}
}

// Add a recipe to the collection.
func (p *Recipes) Add(name string, data []byte) error {
	r, err := Load(data)
	if err != nil {
		return err
	}

	p.recipes[name] = r

	return nil
}

// AddFromPath adds a recipe from a file path.
func (p *Recipes) AddFromPath(path string) error {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return p.Add(name, data)
}

func (p *Recipes) Get(name string) *Recipe {
	return p.recipes[name]
}
