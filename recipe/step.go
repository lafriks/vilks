// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package recipe

import (
	"fmt"
	"strconv"
)

type EvidenceType string

const (
	EvidenceTypeFile   EvidenceType = "file"
	EvidenceTypeOutput EvidenceType = "output"
)

type Evidence struct {
	Name   string       `json:"name"`
	Type   EvidenceType `json:"type"`
	Path   string       `json:"path,omitempty"`
	Regexp string       `json:"regexp,omitempty"`
}

type When struct {
	Status string `json:"status"`
}

type Conditions struct {
	SuccessRegexp string `json:"success_regexp,omitempty"`
	FailureRegexp string `json:"failure_regexp,omitempty"`
}

type Step struct {
	Name        string         `json:"name"`
	Image       string         `json:"image"`
	Environment map[string]any `json:"environment"`
	Conditions  *Conditions    `json:"conditions,omitempty"`
	Commands    []string       `json:"commands"`
	Evidence    []Evidence     `json:"evidence"`
	When        *When          `json:"when,omitempty"`
}

func (s *Step) Environ(params map[string]string) []string {
	env := make([]string, 0, len(s.Environment))

	for k, v := range s.Environment {
		var val string
		switch v := v.(type) {
		case string:
			val = v
		case int, int32, int64:
			val = fmt.Sprintf("%d", v)
		case bool:
			val = strconv.FormatBool(v)
		case float32, float64:
			val = fmt.Sprintf("%f", v)
		case map[string]any:
			if prm, ok := v["from_param"]; ok {
				if v, ok := prm.(string); ok {
					val = params[v]
				}
			}
			if prm, ok := v["from_evidence"]; ok {
				if v, ok := prm.(string); ok {
					val = params["evidence_"+v]
				}
			}
		}

		env = append(env, fmt.Sprintf("%s=%s", k, val))
	}

	return env
}

type Service struct {
	Name    string        `json:"name"`
	Image   string        `json:"image"`
	Command string        `json:"command"`
	Ports   []ServicePort `json:"ports"`
}

type ServicePort struct {
	Name string `json:"name"`
	Port string `json:"port"`
}
