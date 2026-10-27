package process

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func init() {
	specialHelper = func(mode string) bool {
		switch mode {
		case "pty-session":
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			signalCtx, cancelSignal := context.WithCancelCause(ctx)
			defer cancelSignal(nil)
			executable, _ := os.Executable()
			e := &Executor{Dir: os.Getenv("UPMACX_PTY_DIR"), In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Grace: 20 * time.Millisecond, OnSignal: func(s syscall.Signal) { cancelSignal(SignalError{Signal: s}) }}
			_, err := e.Run(signalCtx, Command{Path: executable, Args: []string{"-test.run=^TestHelper$", "--", "pty-child"}, Mode: "stream", Stdin: "inherit"})
			if cause, ok := context.Cause(signalCtx).(SignalError); ok {
				fmt.Println("CHILD_SIGNAL:" + cause.Signal.String())
			} else if err != nil {
				fmt.Println("SESSION_ERROR:" + err.Error())
				os.Exit(9)
			}
			var foreground int32
			_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCGPGRP, uintptr(unsafe.Pointer(&foreground)))
			if errno != 0 || foreground != int32(syscall.Getpgrp()) {
				fmt.Println("FOREGROUND_NOT_RESTORED")
				os.Exit(10)
			}
			fmt.Println("FOREGROUND_RESTORED")
			return true
		case "pty-child":
			fmt.Printf("CHILD_READY:%d\n", os.Getpid())
			line, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				os.Exit(11)
			}
			if strings.TrimSpace(line) != "guest-test-input" {
				os.Exit(12)
			}
			fmt.Println("CHILD_READ_OK")
			return true
		}
		return false
	}
}

func openPTY(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { master.Close() })
	for _, request := range []uintptr{syscall.TIOCPTYGRANT, syscall.TIOCPTYUNLK} {
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), request, 0); errno != 0 {
			t.Fatal(errno)
		}
	}
	var name [128]byte
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCPTYGNAME, uintptr(unsafe.Pointer(&name[0]))); errno != 0 {
		t.Fatal(errno)
	}
	slave, err := os.OpenFile(strings.TrimRight(string(name[:]), "\x00"), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { slave.Close() })
	return master, slave
}

func TestTerminalHandoff(t *testing.T) {
	for _, interrupt := range []bool{false, true} {
		name := "input"
		if interrupt {
			name = "ctrl-c"
		}
		t.Run(name, func(t *testing.T) {
			master, slave := openPTY(t)
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			child := exec.Command(exe, "-test.run=^TestHelper$", "--", "pty-session")
			child.Env = append(os.Environ(), "UPMACX_FAKE=1", "UPMACX_PTY_DIR="+dir, "PATH="+dir, "HOME="+dir, "GORACE=atexit_sleep_ms=0")
			child.Stdin = slave
			child.Stdout = slave
			child.Stderr = slave
			child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
			if err = child.Start(); err != nil {
				t.Fatal(err)
			}
			slave.Close()
			var childGroup int
			t.Cleanup(func() {
				if childGroup > 0 {
					_ = syscall.Kill(-childGroup, syscall.SIGKILL)
				}
				_ = syscall.Kill(-child.Process.Pid, syscall.SIGKILL)
			})
			lines := make(chan string, 32)
			go func() {
				defer close(lines)
				scanner := bufio.NewScanner(master)
				for scanner.Scan() {
					lines <- strings.TrimSpace(scanner.Text())
				}
			}()
			done := make(chan error, 1)
			go func() { done <- child.Wait() }()
			timeout := time.After(10 * time.Second)
			var transcript []string
			sent := false
			restored := false
			observed := false
			for !restored {
				select {
				case line, ok := <-lines:
					if !ok {
						t.Fatalf("terminal closed: %v", transcript)
					}
					transcript = append(transcript, line)
					if strings.HasPrefix(line, "CHILD_READY:") {
						pid, err := strconv.Atoi(strings.TrimPrefix(line, "CHILD_READY:"))
						if err != nil {
							t.Fatal(err)
						}
						childGroup = pid
						if sent {
							t.Fatal("duplicate child prompt")
						}
						sent = true
						if interrupt {
							_, err = master.Write([]byte{3})
						} else {
							_, err = master.WriteString("guest-test-input\n")
						}
						if err != nil {
							t.Fatal(err)
						}
					}
					if line == "FOREGROUND_RESTORED" {
						restored = true
					}
					if !interrupt && line == "CHILD_READ_OK" || interrupt && strings.HasSuffix(line, "CHILD_SIGNAL:interrupt") {
						observed = true
					}
				case <-timeout:
					t.Fatalf("terminal test timed out: %v", transcript)
				}
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("session failed: %v, %v", err, transcript)
				}
			case <-timeout:
				t.Fatal("terminal session did not exit")
			}
			if !sent || !observed {
				t.Fatalf("missing terminal interaction: %v", transcript)
			}
		})
	}
}
