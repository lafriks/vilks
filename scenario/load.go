// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package scenario

import (
	"context"
	"os"

	"github.com/goccy/go-yaml"
)

func Load(ctx context.Context, data []byte) (*Scenario, error) {
	s := &Scenario{}

	if err := yaml.UnmarshalContext(ctx, data, s); err != nil {
		return nil, err
	}

	return s, nil
}

func LoadFile(ctx context.Context, path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Load(ctx, data)
}
