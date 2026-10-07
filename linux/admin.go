package main

import (
	"os"
	"os/exec"
)

// IsAdmin checks if the current process is running with root privileges on Linux (UID 0).
func IsAdmin() bool {
	return os.Geteuid() == 0
}

// RelaunchAsAdmin attempts to restart the current process with sudo on Linux.
func RelaunchAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command("sudo", append([]string{exe}, os.Args[1:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
