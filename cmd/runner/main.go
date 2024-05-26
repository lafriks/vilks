// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"fmt"
	"os"

	"vilks.io/vilks/logger"

	"github.com/spf13/cobra"
)

var Version = "0.1.0-dev"

var RootCmd *cobra.Command

var log logger.Logger

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func initRootCmd() {
	if RootCmd != nil {
		return
	}

	log = &logger.Console{}

	RootCmd = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			debug, _ := cmd.PersistentFlags().GetBool("debug")
			log.SetDebug(debug)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	RootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
}

func main() {
	initRootCmd()

	RootCmd.Version = Version

	Execute()
}
