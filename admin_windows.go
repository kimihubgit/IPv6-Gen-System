//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// IsAdmin checks if the current process is running with Administrator privileges on Windows.
// On Windows, the command `net session` returns exit code 0 only when run as Administrator.
func IsAdmin() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// RelaunchAsAdmin attempts to restart the current process with elevated privileges (UAC prompt).
func RelaunchAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}

	args := os.Args[1:]
	quotedArgs := make([]string, len(args))
	for i, a := range args {
		quotedArgs[i] = fmt.Sprintf(`'%s'`, a)
	}
	argsStr := strings.Join(quotedArgs, ", ")

	// PowerShell Start-Process -Verb RunAs
	var psCmd string
	if len(quotedArgs) > 0 {
		psCmd = fmt.Sprintf(`Start-Process -FilePath '%s' -ArgumentList @(%s) -WorkingDirectory '%s' -Verb RunAs`, exe, argsStr, cwd)
	} else {
		psCmd = fmt.Sprintf(`Start-Process -FilePath '%s' -WorkingDirectory '%s' -Verb RunAs`, exe, cwd)
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	return cmd.Start()
}
