package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

var gitExecutable string

func init() {
	var err error
	gitExecutable, err = exec.LookPath("git")
	if err != nil {
		log.Fatalf("could not locate git: '%#v'", err)
	}
}

// Git calls the system git in the project directory with specified arguments
func Git(writer io.Writer, args ...string) (int, error) {
	command := exec.Command(gitExecutable, args...)
	command.Stdout = writer
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	return StartAndWait(command)
}

// StartAndWait calls the command and returns the result
func StartAndWait(command *exec.Cmd) (int, error) {
	var err error
	if err = command.Start(); err != nil {
		return -1, fmt.Errorf("could not start command: %w", err)
	}
	err = command.Wait()
	if err == nil {
		return 0, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		if err.(*exec.ExitError).Stderr == nil {
			return 0, nil
		}
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), fmt.Errorf("error waiting for command: %w", err)
		}
	} else {
		return -1, fmt.Errorf("error waiting for command: %w", err)
	}
	return 0, nil
}
