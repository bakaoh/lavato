package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// KillProcess sends KILLINT signal command to the given process
func KillProcess(processName string) error {

	pids, err := resolvePids(processName)
	if err != nil {
		return err
	}
	if len(pids) == 0 {
		return fmt.Errorf("no %s processes running", processName)
	}
	if len(pids) > 1 {
		errStr := fmt.Sprintf("multiple %s processes running:\n", processName)
		prefix := ""
		for _, p := range pids {
			errStr += fmt.Sprintf("%s%d", prefix, p)
			prefix = "\n"
		}
		return errors.New(errStr)
	}
	pid := pids[0]

	return kill(pid, syscall.SIGINT)
}

// resolvePids returns the pids for all running gnatsd processes.
func resolvePids(processName string) ([]int, error) {
	// If pgrep isn't available, this will just bail out and the user will be
	// required to specify a pid.
	output, err := pgrep(processName)
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// ExitError indicates non-zero exit code, meaning no processes found.
			break
		default:
			return nil, errors.New("unable to resolve pid, try providing one")
		}
	}
	var (
		myPid   = os.Getpid()
		pidStrs = strings.Split(string(output), "\n")
		pids    = make([]int, 0, len(pidStrs))
	)
	for _, pidStr := range pidStrs {
		if pidStr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return nil, errors.New("unable to resolve pid, try providing one")
		}
		// Ignore the current process.
		if pid == myPid {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

var kill = func(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}

var pgrep = func(processName string) ([]byte, error) {
	return exec.Command("pgrep", "-f", processName).Output()
}
