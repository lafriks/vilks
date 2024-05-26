// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package logger

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	SetDebug(debug bool)
	Info(msg string, params ...any)
	Error(msg string)
	Debug(msg string)
	Console(title string, data []byte)
	Special(data string) string
}
