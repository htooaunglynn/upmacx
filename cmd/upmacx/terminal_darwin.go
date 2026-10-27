//go:build darwin

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func isTerminal() bool {
	var t syscall.Termios
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCGETA, uintptr(unsafe.Pointer(&t)))
	return e == 0
}
