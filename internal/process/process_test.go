package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

var specialHelper = func(string) bool { return false }

func TestHelper(t *testing.T) {
	if os.Getenv("UPMACX_FAKE") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		os.Exit(3)
	}
	if specialHelper(args[1]) {
		os.Exit(0)
	}
	switch args[1] {
	case "echo":
		dir, _ := os.Getwd()
		_ = json.NewEncoder(os.Stdout).Encode([]string{dir, os.Getenv("UPMACX_VALUE"), strings.Join(args[2:], "|")})
		fmt.Fprint(os.Stderr, "fake stderr")
	case "fail":
		os.Exit(7)
	case "signal":
		_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
		time.Sleep(time.Hour)
	case "tree", "signal-tree":
		child := exec.Command(os.Args[0], "-test.run=^TestHelper$", "--", "descendant")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		child.Env = os.Environ()
		if child.Start() != nil {
			os.Exit(4)
		}
		if args[1] == "signal-tree" {
			deadline := time.Now().Add(3 * time.Second)
			for {
				if _, err := os.Stat(os.Getenv("UPMACX_READY")); err == nil {
					break
				}
				if time.Now().After(deadline) {
					os.Exit(6)
				}
				time.Sleep(5 * time.Millisecond)
			}
			_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
		}
		_ = child.Wait()
	case "descendant":
		signal.Ignore(syscall.SIGTERM, syscall.SIGINT)
		_ = os.WriteFile(os.Getenv("UPMACX_READY"), []byte("ready"), 0600)
		for {
			f, e := os.OpenFile(os.Getenv("UPMACX_HEARTBEAT"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if e != nil {
				os.Exit(5)
			}
			_, _ = f.Write([]byte("."))
			_ = f.Close()
			time.Sleep(10 * time.Millisecond)
		}
	case "hang":
		time.Sleep(time.Hour)
	case "overflow":
		_, _ = io.CopyN(os.Stdout, zeroReader{}, Limit+1)
	}
	os.Exit(0)
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }
func setup(t *testing.T) (*Executor, Command) {
	t.Helper()
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	e2 := &Executor{Dir: dir, Out: io.Discard, Err: io.Discard, Grace: 20 * time.Millisecond}
	return e2, Command{Path: exe, Args: []string{"-test.run=^TestHelper$", "--", "echo"}, Env: map[string]string{"UPMACX_FAKE": "1", "PATH": dir, "HOME": dir, "UPMACX_VALUE": "safe", "GORACE": "atexit_sleep_ms=0"}, Mode: "capture", Stdin: "closed"}
}
func TestCapture(t *testing.T) {
	e, c := setup(t)
	c.Args = append(c.Args, "a;$(false)", "space name")
	r, err := e.Run(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	var values []string
	if json.Unmarshal(r.Stdout, &values) != nil || values[0] != canonical(e.Dir) || values[1] != "safe" || values[2] != "a;$(false)|space name" {
		t.Fatal(string(r.Stdout))
	}
	if string(r.Stderr) != "fake stderr" {
		t.Fatal(r)
	}
}
func TestProcessErrors(t *testing.T) {
	for _, mode := range []string{"fail", "overflow", "hang", "missing"} {
		t.Run(mode, func(t *testing.T) {
			e, c := setup(t)
			c.Args[len(c.Args)-1] = mode
			if mode == "missing" {
				c.Path = filepath.Join(e.Dir, "missing")
			}
			e.CaptureTimeout = 5 * time.Second
			if mode == "hang" {
				e.CaptureTimeout = 200 * time.Millisecond
			}
			r, err := e.Run(context.Background(), c)
			if err == nil {
				t.Fatal("expected error")
			}
			if mode == "fail" && r.ExitCode != 7 {
				t.Fatal(r)
			}
			if mode == "overflow" && !errors.Is(err, ErrOverflow) {
				t.Fatal(err)
			}
			if mode == "hang" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
		})
	}
}
func TestCancellation(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		e, c := setup(t)
		c.Args[len(c.Args)-1] = "hang"
		ctx, cancel := context.WithCancelCause(context.Background())
		timer := time.AfterFunc(50*time.Millisecond, func() { cancel(SignalError{Signal: sig}) })
		start := time.Now()
		_, err := e.Run(ctx, c)
		timer.Stop()
		cancel(nil)
		if !errors.Is(err, context.Canceled) || time.Since(start) > time.Second*3 {
			t.Fatal(err)
		}
	}
}
func TestStream(t *testing.T) {
	e, c := setup(t)
	var out, errout strings.Builder
	e.Out = &out
	e.Err = &errout
	c.Mode = "stream"
	if _, err := e.Run(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 || errout.String() != "fake stderr" {
		t.Fatal("missing stream output")
	}
}

func canonical(p string) string {
	v, e := filepath.EvalSymlinks(p)
	if e != nil {
		return p
	}
	return v
}

func TestChildSignal(t *testing.T) {
	e, c := setup(t)
	c.Args[len(c.Args)-1] = "signal"
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	e.OnSignal = func(s syscall.Signal) { cancel(SignalError{Signal: s}) }
	_, err := e.Run(ctx, c)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if s, ok := context.Cause(ctx).(SignalError); !ok || s.Signal != syscall.SIGINT {
		t.Fatal(context.Cause(ctx))
	}
}
func TestCancellationKillsDescendant(t *testing.T) {
	e, c := setup(t)
	timerUsed := false
	e.NewTimer = func(time.Duration) (<-chan time.Time, func()) {
		timerUsed = true
		ch := make(chan time.Time)
		close(ch)
		return ch, func() {}
	}
	c.Args[len(c.Args)-1] = "tree"
	ready := filepath.Join(e.Dir, "ready")
	heartbeat := filepath.Join(e.Dir, "heartbeat")
	c.Env["UPMACX_READY"] = ready
	c.Env["UPMACX_HEARTBEAT"] = heartbeat
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	done := make(chan error, 1)
	go func() { _, err := e.Run(ctx, c); done <- err }()
	deadline := time.After(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("descendant did not start")
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel(SignalError{Signal: syscall.SIGTERM})
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process group did not exit")
	}
	if !timerUsed {
		t.Fatal("grace timer was not used")
	}
	first, err := os.ReadFile(heartbeat)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(60 * time.Millisecond)
	second, err := os.ReadFile(heartbeat)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(second) {
		t.Fatal("descendant survived cancellation")
	}
}

func TestLeaderExitPreservesDescendantGrace(t *testing.T) {
	e, c := setup(t)
	c.Args[len(c.Args)-1] = "tree"
	ready := filepath.Join(e.Dir, "ready")
	heartbeat := filepath.Join(e.Dir, "heartbeat")
	c.Env["UPMACX_READY"] = ready
	c.Env["UPMACX_HEARTBEAT"] = heartbeat
	// Files avoid Wait's output-copy pipes, so the leader can be reaped while
	// its signal-ignoring descendant is still alive.
	output, err := os.Create(filepath.Join(e.Dir, "output"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	e.Out = output
	e.Err = output
	c.Mode = "stream"
	ticks := make(chan time.Time)
	started := make(chan struct{})
	e.NewTimer = func(time.Duration) (<-chan time.Time, func()) { close(started); return ticks, func() {} }
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	done := make(chan error, 1)
	go func() { _, err := e.Run(ctx, c); done <- err }()
	deadline := time.After(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("descendant not ready")
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel(SignalError{Signal: syscall.SIGTERM})
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("no grace timer")
	}
	select {
	case <-done:
		t.Fatal("returned before injected grace period ended")
	case <-time.After(60 * time.Millisecond):
	}
	first, err := os.ReadFile(heartbeat)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(60 * time.Millisecond)
	second, err := os.ReadFile(heartbeat)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) <= len(first) {
		t.Fatal("descendant killed before grace period")
	}
	close(ticks)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("group not terminated")
	}
}

func TestInterruptedLeaderWithOpenDescendantPipes(t *testing.T) {
	e, c := setup(t)
	c.Args[len(c.Args)-1] = "signal-tree"
	c.Env["UPMACX_READY"] = filepath.Join(e.Dir, "ready")
	c.Env["UPMACX_HEARTBEAT"] = filepath.Join(e.Dir, "heartbeat")
	e.CaptureTimeout = 5 * time.Second
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	e.OnSignal = func(s syscall.Signal) { cancel(SignalError{Signal: s}) }
	start := time.Now()
	_, err := e.Run(ctx, c)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected signal cancellation, got %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("signal notification waited for inherited pipes")
	}
}
