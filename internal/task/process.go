package task

import (
	"os/exec"
	"syscall"
)

func startCommand(name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd, cmd.Start()
}

func pauseProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(syscall.SIGSTOP)
}

func resumeProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(syscall.SIGCONT)
}

func cancelProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	pgid, _ := syscall.Getpgid(cmd.Process.Pid)
	syscall.Kill(-pgid, syscall.SIGKILL)
	return cmd.Process.Kill()
}
