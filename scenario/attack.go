// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package scenario

type Scenario struct {
	Name  string `json:"name"`
	Teams []Team `json:"teams"`
	Hosts []Host `json:"hosts"`
}

type Team struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type Host struct {
	Name    string   `json:"name"`
	Target  string   `json:"host"`
	Attacks []Attack `json:"attacks"`
}

type Attack struct {
	Name   string  `json:"name"`
	Recipe string  `json:"recipe"`
	Params []Param `json:"params"`
}

type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
