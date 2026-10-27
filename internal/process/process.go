package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Command struct {
	Path             string
	Args             []string
	Env              map[string]string
	Dir, Mode, Stdin string
}
type CommandResult struct {
	Stdout, Stderr []byte
	ExitCode       int
}
type Runner interface {
	LookPath(string) (string, error)
	Run(context.Context, Command) (CommandResult, error)
}
type Executor struct {
	In                    io.Reader
	Out, Err              io.Writer
	Dir                   string
	CaptureTimeout, Grace time.Duration
	OnSignal              func(syscall.Signal)
	NewTimer              func(time.Duration) (<-chan time.Time, func())
}

const Limit = 16 << 20

var ErrOverflow = errors.New("command output exceeded 16 MiB")

type bounded struct {
	b        []byte
	overflow bool
	cancel   context.CancelFunc
}

func (b *bounded) Write(p []byte) (int, error) {
	n := len(p)
	if len(b.b)+n > Limit {
		b.overflow = true
		b.cancel()
		return n, nil
	}
	b.b = append(b.b, p...)
	return n, nil
}
func (e *Executor) LookPath(name string) (string, error) {
	p, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return filepath.Abs(p)
}
func (e *Executor) Run(ctx context.Context, c Command) (CommandResult, error) {
	r := CommandResult{ExitCode: -1}
	if !filepath.IsAbs(c.Path) {
		return r, errors.New("executable path must be absolute")
	}
	if c.Mode != "capture" && c.Mode != "stream" {
		return r, errors.New("invalid command mode")
	}
	if c.Stdin != "closed" && c.Stdin != "inherit" {
		return r, errors.New("invalid stdin policy")
	}
	timeout := e.CaptureTimeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	var cancel context.CancelFunc
	if c.Mode == "capture" {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()
	child := exec.Command(c.Path, c.Args...)
	configure(child)
	child.Dir = c.Dir
	if child.Dir == "" {
		child.Dir = e.Dir
	}
	if child.Dir == "" {
		return r, errors.New("neutral working directory required")
	}
	child.Env = os.Environ()
	for k, v := range c.Env {
		prefix := k + "="
		env := child.Env[:0]
		for _, entry := range child.Env {
			if !strings.HasPrefix(entry, prefix) {
				env = append(env, entry)
			}
		}
		child.Env = append(env, prefix+v)
	}
	out, errout := &bounded{cancel: cancel}, &bounded{cancel: cancel}
	stdout, stderr := e.Out, e.Err
	if c.Mode == "capture" {
		stdout = out
		stderr = errout
	}
	stdoutWriter, stdoutPipe, err := newOutput(stdout)
	child.Stdout = stdoutWriter
	if err != nil {
		return r, fmt.Errorf("stdout pipe: %w", err)
	}
	defer stdoutPipe.close()
	stderrWriter, stderrPipe, err := newOutput(stderr)
	child.Stderr = stderrWriter
	if err != nil {
		return r, fmt.Errorf("stderr pipe: %w", err)
	}
	defer stderrPipe.close()
	if c.Stdin == "inherit" {
		child.Stdin = e.In
	}
	if err := ctx.Err(); err != nil {
		return r, err
	}
	restore := terminalForeground(child, c, e.In)
	defer restore()
	if err := child.Start(); err != nil {
		return r, fmt.Errorf("start command: %w", err)
	}
	stdoutPipe.start(cancel)
	stderrPipe.start(cancel)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-done:
			if ctx.Err() == nil {
				return
			}
		case <-ctx.Done():
		}
		{
			sig := syscall.SIGTERM
			if s, ok := context.Cause(ctx).(SignalError); ok {
				sig = s.Signal
			}
			signalGroup(child, sig)
			grace := e.Grace
			if grace == 0 {
				grace = 5 * time.Second
			}
			var ticks <-chan time.Time
			var stop func()
			if e.NewTimer != nil {
				ticks, stop = e.NewTimer(grace)
			} else {
				timer := time.NewTimer(grace)
				ticks = timer.C
				stop = func() { timer.Stop() }
			}
			defer stop()
			// The leader may exit before descendants. Preserve the full grace
			// period even if Wait finishes, then terminate any remaining group.
			<-ticks
			signalGroup(child, syscall.SIGKILL)
		}
	}()
	err = child.Wait()
	// Foreground terminal signals may reach only the child group. Notify the
	// controller before the watcher exits so descendants are cleaned up too.
	r.ExitCode = child.ProcessState.ExitCode()
	if status, ok := child.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		sig := status.Signal()
		if (sig == syscall.SIGINT || sig == syscall.SIGTERM) && e.OnSignal != nil {
			e.OnSignal(sig)
		}
	}
	stdoutErr := stdoutPipe.wait()
	stderrErr := stderrPipe.wait()
	close(done)
	wg.Wait()
	r.Stdout = out.b
	r.Stderr = errout.b
	if out.overflow || errout.overflow {
		return r, ErrOverflow
	}
	if stdoutErr != nil || stderrErr != nil {
		return r, fmt.Errorf("command output: %w", errors.Join(stdoutErr, stderrErr))
	}
	if ctx.Err() != nil {
		return r, ctx.Err()
	}
	if err != nil {
		return r, fmt.Errorf("command exited %d: %w", r.ExitCode, err)
	}
	return r, nil
}

type SignalError struct{ Signal syscall.Signal }

func (s SignalError) Error() string { return s.Signal.String() }
