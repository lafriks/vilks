// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"vilks.io/vilks/runner"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/client"
	"github.com/moby/moby/pkg/stdcopy"
)

const (
	volumeDriver = "local"
	wordspaceDir = "/workspace"
	evidenceDir  = "/evidence"
)

var (
	startOpts = container.StartOptions{}

	removeOpts = container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	}

	logsOpts = container.LogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	}
)

var ErrContainerNotStarted = errors.New("container not started")

type impl struct {
	containerID        string
	volumeName         string
	evidenceVolumeName string
	client             *client.Client
}

func New() runner.Runner {
	return &impl{}
}

func (d *impl) connect() error {
	if d.client != nil {
		return nil
	}

	c, err := client.NewClientWithOpts()
	if err != nil {
		return err
	}

	d.client = c

	return nil
}

func (d *impl) CreateWorkspace(_ context.Context, dir string) error {
	// TODO: copy files from dir to volume (using virtual volume)
	volName, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	d.volumeName = volName

	return nil
}

func (d *impl) CreateEvidenceStore(_ context.Context, dir string) error {
	volName, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	d.evidenceVolumeName = volName

	return nil
}

func (d *impl) Start(ctx context.Context, cmd runner.StartOptions) error {
	if err := d.connect(); err != nil {
		return err
	}

	entrypoint := cmd.Entrypoint
	if !cmd.Plugin && !cmd.Service && len(entrypoint) == 0 {
		entrypoint = []string{cmd.Shell, "-c", fmt.Sprintf("sleep %d", int(cmd.Timeout.Seconds()))}
	}

	containerConfig := &container.Config{
		Image:      cmd.Image,
		Env:        nil,
		Entrypoint: entrypoint,
	}

	hostConfig := &container.HostConfig{}
	if d.volumeName != "" {
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", d.volumeName, wordspaceDir))
	}

	if d.evidenceVolumeName != "" {
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", d.evidenceVolumeName, evidenceDir))
	}

	if cmd.Service && len(cmd.Ports) > 0 {
		hostConfig.PortBindings = make(nat.PortMap, len(cmd.Ports))
		containerConfig.ExposedPorts = make(nat.PortSet, len(cmd.Ports))

		_, ports, err := nat.ParsePortSpecs(cmd.Ports)
		if err != nil {
			return err
		}

		hostConfig.PortBindings = ports

		for p := range ports {
			containerConfig.ExposedPorts[p] = struct{}{}
		}
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if client.IsErrNotFound(err) {
		var r io.ReadCloser

		r, err = d.client.ImagePull(ctx, cmd.Image, image.PullOptions{})
		if err != nil {
			return err
		}

		// Read all response to wait for the image to be pulled
		_, _ = io.ReadAll(r)
		r.Close()

		resp, err = d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	}

	if err != nil {
		return err
	}

	d.containerID = resp.ID

	return d.client.ContainerStart(ctx, d.containerID, startOpts)
}

func (d *impl) Tail(ctx context.Context) (io.ReadCloser, error) {
	if d.containerID == "" {
		return nil, ErrContainerNotStarted
	}

	if err := d.connect(); err != nil {
		return nil, err
	}

	logs, err := d.client.ContainerLogs(ctx, d.containerID, logsOpts)
	if err != nil {
		return nil, err
	}

	rc, wc := io.Pipe()

	// de multiplex 'logs' who contains two streams, previously multiplexed together using StdWriter
	go func() {
		_, _ = stdcopy.StdCopy(wc, wc, logs)
		_ = logs.Close()
		_ = wc.Close()
	}()

	return rc, nil
}

func (d *impl) Exec(ctx context.Context, env []string, cmd string, args ...string) (*runner.ExecResult, error) {
	if d.containerID == "" {
		return nil, ErrContainerNotStarted
	}

	if err := d.connect(); err != nil {
		return nil, err
	}

	consoleSize := [2]uint{20, 80}

	exec, err := d.client.ContainerExecCreate(ctx, d.containerID, container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   wordspaceDir,
		Env:          env,
		Cmd:          append([]string{cmd}, args...),
	})
	if err != nil {
		return nil, err
	}

	resp, err := d.client.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{
		Detach:      false,
		Tty:         false,
		ConsoleSize: &consoleSize,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// Read the command output
	var outBuf, errBuf bytes.Buffer

	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return nil, err
		}

		break
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	stdout, err := io.ReadAll(&outBuf)
	if err != nil {
		return nil, err
	}

	stderr, err := io.ReadAll(&errBuf)
	if err != nil {
		return nil, err
	}

	res, err := d.client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return nil, err
	}

	return &runner.ExecResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: res.ExitCode,
	}, nil
}

func (d *impl) DownlaodEvidence(ctx context.Context, path string) (io.ReadCloser, error) {
	if d.containerID == "" {
		return nil, ErrContainerNotStarted
	}

	if err := d.connect(); err != nil {
		return nil, err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(wordspaceDir, path)
	}

	rc, _, err := d.client.CopyFromContainer(ctx, d.containerID, path)

	return rc, err
}

func (d *impl) Stop(ctx context.Context) error {
	if d.containerID != "" {
		if err := d.client.ContainerKill(ctx, d.containerID, "9"); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
			return err
		}

		if err := d.client.ContainerRemove(ctx, d.containerID, removeOpts); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
			return err
		}
	}

	if d.volumeName != "" {
		if err := d.client.VolumeRemove(ctx, d.volumeName, true); err != nil && !strings.Contains(err.Error(), "No such volume") {
			return err
		}
	}

	if d.evidenceVolumeName != "" {
		if err := d.client.VolumeRemove(ctx, d.evidenceVolumeName, true); err != nil && !strings.Contains(err.Error(), "No such volume") {
			return err
		}
	}

	return nil
}

func isErrContainerNotFoundOrNotRunning(err error) bool {
	// Error response from daemon: Cannot kill container: ...: No such container: ...
	// Error response from daemon: Cannot kill container: ...: Container ... is not running"
	// Error response from podman daemon: can only kill running containers. ... is in state exited
	// Error: No such container: ...
	return err != nil && (strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running") || strings.Contains(err.Error(), "can only kill running containers"))
}
