package main

import (
	"fmt"
	"github.com/sascha-andres/flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/sascha-andres/gitc/internal"
)

var (
	l                                            = log.New(os.Stdout, "[git-c] ", log.LstdFlags)
	help, add, printCommitMessage, push, verbose bool
	commitMessageFile                            string
	gitExecutable                                string
)

// main you know
func main() {
	gitHookIntegration()

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	if len(commitMessageFile) > 0 {
		lintCommitMessage(commitMessageFile)
	} else {
		buildCommitMessage()
	}
}

// gitHookIntegration checks whether this is running as a git hook
func gitHookIntegration() {
	args := os.Args
	_, name := path.Split(args[0])
	if name == "commit-msg" {
		if len(args) == 1 {
			l.Print("not enough arguments provided")
			os.Exit(1)
		}
		lintCommitMessage(args[1])
		os.Exit(0)
	}
}

// lintCommitMessage is used to line a message
func lintCommitMessage(file string) {
	data, err := ioutil.ReadFile(file)
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

// buildCommitMessage is used to build a commit message
func buildCommitMessage() {
	cmb := internal.CommitMessageBuilder{}
	if verbose {
		l.Print("asking for data")
	}
	err := cmb.Build()
	if err != nil {
		l.Printf("error creating commit message: %s", err)
		os.Exit(1)
	}
	msg := cmb.String()
	if verbose {
		l.Print("commit message created")
	}

	if printCommitMessage {
		l.Println("resulting commit message:")
		for _, s := range strings.Split(msg, " \n") {
			l.Printf("  %s", s)
		}
	}

	if add {
		if verbose {
			l.Print("stage files")
		}
		Git("add", "--all", ":/")
	}
	if verbose {
		l.Print("commit")
	}
	Git("commit", "-m", msg)
	if push {
		if verbose {
			l.Print("push")
		}
		Git("push")
	}
}

// init is known
func init() {
	var err error
	gitExecutable, err = exec.LookPath("git")
	if err != nil {
		l.Fatalf("could not locate git: '%#v'", err)
	}
	flag.SetEnvPrefix("GIT_C")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&verbose, "verbose", false, "print more information on execution")
	flag.BoolVar(&add, "add", false, "add all changed files before committing")
	flag.BoolVar(&push, "push", false, "automatically push to default remote")
	flag.BoolVar(&printCommitMessage, "print", false, "print generated commit message")
	flag.StringVar(&commitMessageFile, "lint", "", "print generated commit message")
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
