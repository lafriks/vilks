// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package logger

import (
	"fmt"

	"github.com/fatih/color"
)

type Console struct {
	debug bool
}

func (c *Console) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Console) Info(msg string, params ...any) {
	fmt.Println(color.GreenString("[+]"), color.BlueString(msg))

	if len(params) > 0 {
		for _, p := range params {
			switch p := p.(type) {
			case map[string]string:
				for k, v := range p {
					color.White("    > %s: %s", color.BlueString(k), color.CyanString(v))
				}
			case string:
				color.White("    > %s", color.CyanString(p))
			}
		}
	}
}

func (c *Console) Special(data string) string {
	return color.MagentaString(data)
}

func (c *Console) Error(msg string) {
	color.HiRed("[!] %s", color.RedString(msg))
}

func (c *Console) Debug(msg string) {
	if !c.debug {
		return
	}

	color.White("[*] %s", msg)
}

func (c *Console) Console(title string, data []byte) {
	if !c.debug {
		return
	}

	if len(data) == 0 {
		color.White("[*] %s: %s", title, color.HiWhiteString("no output"))

		return
	}

	color.White("[*] %s", title)
	color.White("-------------------------")
	fmt.Println(string(data))
	color.White("-------------------------")
}

var _ Logger = &Console{}
