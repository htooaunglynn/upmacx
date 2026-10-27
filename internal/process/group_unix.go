//go:build darwin || linux

package process

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"
)

func configure(c *exec.Cmd)                     { c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }
func signalGroup(c *exec.Cmd, s syscall.Signal) { _ = syscall.Kill(-c.Process.Pid, s) }

// Put interactive installers in the terminal foreground while keeping a distinct
// process group, so native authentication can read the controlling terminal.
func terminalForeground(child *exec.Cmd, command Command, input interface{}) func() {
	if command.Mode != "stream" || command.Stdin != "inherit" {
		return func() {}
	}
	f, ok := input.(*os.File)
	if !ok {
		return func() {}
	}
	var previous int32
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCGPGRP, uintptr(unsafe.Pointer(&previous)))
	if errno != 0 {
		return func() {}
	}
	child.SysProcAttr.Foreground = true
	child.SysProcAttr.Ctty = int(f.Fd())
	return func() {
		ignored := signal.Ignored(syscall.SIGTTOU)
		signal.Ignore(syscall.SIGTTOU)
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCSPGRP, uintptr(unsafe.Pointer(&previous)))
		if !ignored {
			signal.Reset(syscall.SIGTTOU)
		}
	}
}
