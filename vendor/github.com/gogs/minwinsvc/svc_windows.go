// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.package service

//+build windows

package minwinsvc

import (
	"os"
	"sync"

	"golang.org/x/sys/windows/svc"
)

var (
	onExit func()
	guard  sync.Mutex
)

func init() {
	interactive, err := svc.IsAnInteractiveSession()
	if err != nil {
		panic(err)
	}
	// While run as Windows service, it is not an interactive session,
	// but we don't want hook execute to be treated as service, e.g. gogs.exe hook pre-receive.
	if interactive || len(os.Getenv("SSH_ORIGINAL_COMMAND")) > 0 {
		return
	}
	go func() {
		_ = svc.Run("", runner{})

		guard.Lock()
		f := onExit
		guard.Unlock()

		// Don't hold this lock in user code.
		if f != nil {
			f()
		}
		// Make sure we exit.
		os.Exit(0)
	}()
}

func setOnExit(f func()) {
	guard.Lock()
	onExit = f
	guard.Unlock()
}

type runner struct{}

func (runner) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			return false, 0
		}
	}

	return false, 0
}
