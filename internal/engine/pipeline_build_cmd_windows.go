//go:build windows

package engine

import "os/exec"

func configureBuildCmdProc(cmd *exec.Cmd) {
	// Windows has no POSIX process groups; terminate the root process.
}

func killBuildCmdProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
