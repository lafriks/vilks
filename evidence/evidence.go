// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package evidence

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"
)

type Evidence interface {
	AddEvidence(name, typ string, data []byte) error
}

func New(baseDir string) *Manager {
	return &Manager{
		baseDir: baseDir,
	}
}

type Manager struct {
	baseDir string
}

func (m *Manager) Attack(teamName, hostName string) Evidence {
	dir := filepath.Join(m.baseDir, teamName, hostName)
	_ = os.MkdirAll(dir, 0o755)

	return &inst{
		baseDir: dir,
	}
}

type inst struct {
	baseDir string
}

func (i *inst) AddEvidence(name, typ string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	exts, err := mime.ExtensionsByType(typ)
	if err != nil {
		return err
	}

	if len(exts) == 0 {
		return fmt.Errorf("unknown mime type: %s", typ)
	}

	t := time.Now().Format("20060102150405")

	return os.WriteFile(filepath.Join(i.baseDir, t+"_"+name+exts[0]), data, 0o600)
}
