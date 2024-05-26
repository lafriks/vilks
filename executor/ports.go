// Copyright 2024 Lauris BH, Janis Janusjavics. All rights reserved.
// SPDX-License-Identifier: GPL-3.0

package executor

import (
	"errors"
	"net"
)

func assignFreePort() (int, error) {
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err == nil {
		var l *net.TCPListener

		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()

			addr, ok := l.Addr().(*net.TCPAddr)
			if ok {
				return addr.Port, nil
			}
		}
	}

	return 0, errors.New("free port can not be assigned")
}
