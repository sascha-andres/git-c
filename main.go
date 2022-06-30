package main

import (
	"flag"
	"fmt"
	"github.com/sascha-andres/gitc/internal"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	l                             = log.New(os.Stdout, "[git-c] ", log.LstdFlags)
	help, add, printCommitMessage bool
	commitMessageFile             string
	gitExecutable                 string
)

func LookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		return strings.ToLower(val) == "true"
	}
	return defaultVal
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	if len(commitMessageFile) > 0 {
		lintCommitMessage()
	} else {
		buildCommitMessage()
	}
}

func lintCommitMessage() {
	data, err := ioutil.ReadFile(commitMessageFile)
	if err != nil {
		l.Printf("error reading commit message file: %s", err)
		os.Exit(100)
	}
	cml, _ := internal.NewCommitMessageLinter(string(data))
	err = cml.Lint()
	if err != nil {
		l.Printf("linting failed: %s", err)
		os.Exit(1)
	}
}

func buildCommitMessage() {
	cmb := internal.CommitMessageBuilder{}
	err := cmb.Build()
	if err != nil {
		l.Printf("error creating commit message: %s", err)
		os.Exit(1)
	}
	msg := cmb.String()

	if printCommitMessage {
		l.Println("resulting commit message:")
		for _, s := range strings.Split(msg, " \n") {
			l.Printf("  %s", s)
		}
	}

	if add {
		Git("add", "--all", ":/")
	}
	Git("commit", "-m", msg)
}

func init() {
	var err error
	gitExecutable, err = exec.LookPath("git")
	if err != nil {
		l.Fatalf("could not locate git: '%#v'", err)
	}
	flag.BoolVar(&help, "help", LookupEnvOrBool("GIT_C_HELP", false), "show help")
	flag.BoolVar(&add, "add", LookupEnvOrBool("GIT_C_ADD", false), "add all changed files before committing")
	flag.BoolVar(&printCommitMessage, "print", LookupEnvOrBool("GIT_C_PRINT", false), "print generated commit message")
	flag.StringVar(&commitMessageFile, "lint", LookupEnvOrString("GIT_C_LINT", ""), "print generated commit message")
}

// Git calls the system git in the project directory with specified arguments
func Git(args ...string) (int, error) {
	command := exec.Command(gitExecutable, args...)
	command.Stdout = os.Stdout
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
