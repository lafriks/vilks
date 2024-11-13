// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package runner

import (
	"context"
	"io"
	"time"
)

type Runner interface {
	CreateWorkspace(ctx context.Context, dir string) error
	CreateEvidenceStore(ctx context.Context, dir string) error
	Start(ctx context.Context, cmd StartOptions) error
	Tail(ctx context.Context) (io.ReadCloser, error)
	Exec(ctx context.Context, env []string, cmd string, args ...string) (*ExecResult, error)
	DownlaodEvidence(ctx context.Context, path string) (io.ReadCloser, error)
	Stop(ctx context.Context) error
}

type StartOptions struct {
	Image      string
	Plugin     bool
	Service    bool
	Shell      string
	Entrypoint []string
	Ports      []string
	Timeout    time.Duration
}

type ExecResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}
