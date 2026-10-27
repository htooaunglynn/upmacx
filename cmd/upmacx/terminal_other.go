//go:build !darwin

package main

func isTerminal() bool { return false }
