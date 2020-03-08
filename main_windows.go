/*
Copyright (c) 2020 Kane York
Copyright 2012 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

//+build windows

// Based on file: golang.org/x/sys/windows/svc/example/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

var fset = flag.NewFlagSet("", flag.ContinueOnError)
var fsetOverride = flag.NewFlagSet("", flag.ContinueOnError)

var listenAddr = fset.String("l", "127.0.103.111:80", "ipv4 listen address")
var destination = fset.String("d", "", "redirect destination for go links")
var destGolinks = fset.Bool("golinks", false, "use www.golinks.io")

func usage(errmsg string) {
	fmt.Fprint(os.Stderr, getUsage(errmsg))
	os.Exit(2)
}

func getUsage(errmsg string) string {
	return fmt.Sprintf("%s\n\n"+
		"usage: %s -d DESTINATION [-l LISTENADDR] <command>\n"+
		"     where <command> is one of\n"+
		"       install, remove, debug, start, stop, pause or continue;\n"+
		"     DESTINATION is\n"+
		"       the HTTP address of your go link server;\n"+
		"     LISTENADDR is\n"+
		"       a host:port pair to listen on.\n"+
		"The values of DESTINATION and LISTENADDR will be saved when 'install'ing.\n",
		errmsg, os.Args[0])
}

func main() {
	const svcName = "go-redirect-agent"

	err := fset.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		usage("help:")
	}
	if *destGolinks {
		*destination = "https://www.golinks.io/"
	}
	Destination = *destination
	ListenAddr = *listenAddr

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}

	if !isIntSess {
		runService(svcName, false)
		return
	}

	if fset.NArg() < 1 {
		usage("no command specified")
	}

	cmd := strings.ToLower(fset.Arg(0))
	switch cmd {
	case "debug":
		if *destination == "" {
			usage("no destination specified")
		}
		runService(svcName, true)
		return
	case "install":
		err = installService(svcName, "go-redirect-agent")
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		err = controlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = controlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
}

func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func installService(name, desc string) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}
	s, err = m.CreateService(name, exepath, mgr.Config{
		DisplayName: "go/link Redirect Agent",
		Description: fmt.Sprintf("Redirect handler for http://go/ links. https://github.com/riking/go-redirect-agent/\nConfigured for: %q", *destination),
		StartType:   mgr.StartAutomatic,
	}, "-l", *listenAddr, "-d", *destination)
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

var elog debug.Log

type myservice struct{}

// go-redirect-agent config
var ListenAddr = "127.0.103.111:80"
var Destination = "https://example.com/go/"

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	slowtick := time.Tick(2 * time.Second)
	tick := slowtick

	// BEGIN go-redirect-agent
	// Intentionally re-parse flags to allow for overrides in the service config
	err := fset.Parse(args)
	if err != flag.ErrHelp {
		elog.Error(1, fmt.Sprintf("invalid service parameter flags: %v\ncontinuing.", err))
	}
	if *destination != "" {
		Destination = *destination
	}
	if *listenAddr != "" {
		ListenAddr = *listenAddr
	}

	// Start the HTTP server
	ctx := context.Background()
	svr := &http.Server{}
	httpErrors := make(chan error)

	Setup(nil, Destination)
	l, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		elog.Error(2, fmt.Sprintf("could not listen on address: %v", err))
		errno = 1
		return
	}

	go func() {
		for {
			httpErrors <- svr.Serve(l)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	elog.Info(1, fmt.Sprintf("Running on %s -> %s.", ListenAddr, Destination))
	// END go-redirect-agent
loop:
	for {
		select {
		case <-tick:
			// pass
		case hErr := <-httpErrors:
			elog.Error(1, fmt.Sprintf("Encountered error while serving: %v", hErr))
			// TODO declare that we've failed...?
			// END go-redirect-agent
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				// BEGIN go-redirect-agent
				elog.Info(1, "Stopping HTTP server.")
				var err error
				{
					ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
					err = svr.Shutdown(ctx)
					cancel()
				}
				if errors.Is(err, context.DeadlineExceeded) {
					elog.Error(1, fmt.Sprintf("Encountered timeout when closing HTTP server."))
					err = svr.Close()
				}
				if err != nil {
					elog.Error(1, fmt.Sprintf("Could not cleanly close http server: %v", err))
				}
				elog.Info(1, "go-redirect-agent stopped.")
				// END go-redirect-agent
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
