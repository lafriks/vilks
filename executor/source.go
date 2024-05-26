// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package executor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func copySource(dst, src string) error {
	var s io.ReadCloser

	proto, _, ok := strings.Cut(src, "://")
	if !ok {
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()

		s = f
	} else {
		switch proto {
		case "http", "https":
			// nolint: gosec, noctx
			resp, err := http.Get(src)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			s = resp.Body
		default:
			return fmt.Errorf("unsupported protocol '%s'", proto)
		}
	}

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	return nil
}
