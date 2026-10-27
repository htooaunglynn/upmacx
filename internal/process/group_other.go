//go:build !darwin && !linux

package process

import (
	"os/exec"
	"syscall"
)

func configure(c *exec.Cmd)                     {}
func signalGroup(c *exec.Cmd, s syscall.Signal) { _ = c.Process.Kill() }

func terminalForeground(child *exec.Cmd, command Command, input interface{}) func() { return func() {} }
