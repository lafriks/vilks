// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package executor

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"vilks.io/vilks/recipe"
	"vilks.io/vilks/runner"
	"vilks.io/vilks/runner/docker"

	"github.com/drone/envsubst"
)

type Attack struct {
	executor *Executor

	Host     string
	Recipe   *recipe.Recipe
	Params   map[string]string
	Evidence map[string]string
}

func (a *Attack) Values() map[string]string {
	prms := make(map[string]string, len(a.Recipe.Params)+1)

	// Add default parameters.
	prms["team_name"] = a.executor.TeamName
	prms["team_index"] = a.executor.TeamIndex
	prms["listener_host"] = a.executor.AttackerHost

	for _, p := range a.Recipe.Params {
		prms[p.Name] = a.Params[p.Name]
		if len(prms[p.Name]) == 0 {
			prms[p.Name] = p.Default
		}
	}

	return prms
}

func (a *Attack) prepareWorkspace(ctx context.Context, r runner.Runner) (string, error) {
	dir, err := os.MkdirTemp("", "vilks-workspace-")
	if err != nil {
		return "", err
	}

	if err := os.Chmod(dir, 0o755); err != nil {
		_ = os.RemoveAll(dir)

		return "", err
	}

	if err := r.CreateWorkspace(ctx, dir); err != nil {
		_ = os.RemoveAll(dir)

		return "", err
	}

	for _, item := range a.Recipe.Workspace {
		if err := copySource(filepath.Join(dir, item.Target), item.Source); err != nil {
			_ = os.RemoveAll(dir)

			return "", err
		}
	}

	return dir, nil
}

func (a *Attack) startServices(ctx context.Context) ([]runner.Runner, map[string]string, error) {
	services := make([]runner.Runner, 0, len(a.Recipe.Services))
	params := make(map[string]string, len(a.Recipe.Services))

	for _, svc := range a.Recipe.Services {
		a.executor.log.Info("Starting service " + a.executor.log.Special(svc.Name))

		r := docker.New()
		ports := make([]string, len(svc.Ports))

		for i, p := range svc.Ports {
			hp, err := assignFreePort()
			if err != nil {
				a.stopServices(ctx, services)

				return nil, nil, err
			}

			a.executor.log.Debug(fmt.Sprintf("Assigning port %d to %s", hp, p.Port))

			params[p.Name] = strconv.FormatInt(int64(hp), 10)
			ports[i] = fmt.Sprintf("%d:%s", hp, p.Port)
		}

		if err := r.Start(ctx, runner.StartOptions{
			Image:      svc.Image,
			Service:    true,
			Ports:      ports,
			Entrypoint: []string{"/bin/sh", "-c", svc.Command},
		}); err != nil {
			a.stopServices(ctx, services)

			return nil, nil, err
		}

		services = append(services, r)
	}

	return services, params, nil
}

func (a *Attack) stopServices(ctx context.Context, services []runner.Runner) {
	for _, s := range services {
		_ = s.Stop(ctx)
	}
}

func (a *Attack) executeStep(ctx context.Context, r runner.Runner, step *recipe.Step, params map[string]string) error {
	if err := r.Start(ctx, runner.StartOptions{
		Image:   step.Image,
		Timeout: 20 * time.Minute,
		Shell:   "/bin/sh",
	}); err != nil {
		return err
	}

	defer func() {
		_ = r.Stop(ctx)
	}()

	for _, cmd := range step.Commands {
		t, err := envsubst.Parse(cmd)
		if err != nil {
			return err
		}

		cmd, err = t.Execute(func(s string) string {
			for k, v := range params {
				if strings.EqualFold(s, k) {
					return v
				}
			}

			return ""
		})
		if err != nil {
			return err
		}

		a.executor.log.Debug("Executing command: " + cmd)

		out, err := r.Exec(ctx, step.Environ(params), "/bin/sh", "-c", cmd)
		if err != nil {
			return err
		}

		if out.ExitCode != 0 {
			// TODO: Better command failure handling.
			return fmt.Errorf("command failed: %s", out.Stderr)
		}

		for _, ev := range step.Evidence {
			switch ev.Type {
			case recipe.EvidenceTypeFile:
				s, err := os.Stat(ev.Path)
				if err != nil {
					return err
				}
				if s.IsDir() {
					return fmt.Errorf("evidence path '%s' must be file", ev.Path)
				}

				// TODO: Mark evidence as file
				a.Evidence[ev.Name] = ev.Path
			case recipe.EvidenceTypeOutput:
				r, err := regexp.Compile(ev.Regexp)
				if err != nil {
					return err
				}

				s := string(r.Find([]byte(out.Stdout)))
				if len(s) == 0 {
					return fmt.Errorf("evidence regexp '%s' did not match any output", ev.Regexp)
				}

				a.Evidence[ev.Name] = s
			}
		}

		if err := a.executor.ev.AddEvidence(step.Name+"_output", "text/plain", out.Stdout); err != nil {
			return err
		}

		a.executor.log.Console("Command output", out.Stdout)
	}

	return nil
}

func (a *Attack) Execute(ctx context.Context) error {
	params := maps.Clone(a.Values())

	params["target_host"] = a.Host

	r := docker.New()

	workspaceDir, err := a.prepareWorkspace(ctx, r)
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspaceDir)

	services, prms, err := a.startServices(ctx)
	if err != nil {
		return err
	}
	defer a.stopServices(ctx, services)

	for k, v := range prms {
		// Do not overwrite existing parameters.
		_, ok := params[k]
		if ok {
			a.executor.log.Debug(fmt.Sprintf("Skipping parameter '%s' as this would override provided parameter", k))

			continue
		}

		params[k] = v
	}

	for _, step := range a.Recipe.Steps {
		a.executor.log.Debug("Executing step " + a.executor.log.Special(step.Name))

		prms := maps.Clone(params)

		// Add evidence parameters.
		for k, v := range a.Evidence {
			// TODO: Add _file suffix for file evidence.
			prms["evidence_"+k] = v
		}

		if err := a.executeStep(ctx, r, step, prms); err != nil {
			return err
		}
	}

	return nil
}
