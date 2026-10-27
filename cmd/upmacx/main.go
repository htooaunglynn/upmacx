package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/htooaunglynn/upmacx/internal/cli"
	"github.com/htooaunglynn/upmacx/internal/process"
	"github.com/htooaunglynn/upmacx/internal/providers/brew"
	"github.com/htooaunglynn/upmacx/internal/updater"
)

var version = "dev"

func main() { os.Exit(run()) }
func run() int {
	// Validate and answer informational invocations before filesystem setup.
	opts, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	app := cli.App{Version: version, OS: runtime.GOOS, UID: os.Geteuid(), In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Terminal: isTerminal}
	if opts.Help || opts.Version || runtime.GOOS != "darwin" || os.Geteuid() == 0 {
		return app.Run(context.Background(), os.Args[1:])
	}
	dir, err := os.MkdirTemp("", "upmacx-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case s := <-signals:
			cancel(process.SignalError{Signal: s.(syscall.Signal)})
		case <-done:
		}
	}()
	runner := &process.Executor{In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Dir: dir, OnSignal: func(s syscall.Signal) { cancel(process.SignalError{Signal: s}) }}
	app.Providers = []updater.Provider{brew.New(runner)}
	return app.Run(ctx, os.Args[1:])
}
