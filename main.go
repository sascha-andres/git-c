package main

import (
	"fmt"
	"github.com/sascha-andres/flag"
	"github.com/sascha-andres/gitc/internal/builder"
	"github.com/sascha-andres/gitc/internal/linter"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

var (
	help, add, printCommitMessage, push, verbose bool
	subjectLineLength, bodyLineLength            int
	commitMessageFile                            string
	gitExecutable                                string
)

// main you know
func main() {
	gitHookIntegration()

	flag.Parse()
	if verbose {
		log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if len(commitMessageFile) > 0 {
		lintCommitMessage(commitMessageFile)
	} else {
		err := buildCommitMessage()
		if err != nil {
			log.Printf("%s", err)
			os.Exit(1)
		}
	}
}

// gitHookIntegration checks whether this is running as a git hook
func gitHookIntegration() {
	args := os.Args
	_, name := path.Split(args[0])
	if name == "commit-msg" {
		if len(args) == 1 {
			log.Print("not enough arguments provided")
			os.Exit(1)
		}
		lintCommitMessage(args[1])
		os.Exit(0)
	}
}

// lintCommitMessage is used to line a message
func lintCommitMessage(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Printf("error reading commit message file: %s", err)
		os.Exit(100)
	}
	cml, _ := linter.NewCommitMessageLinter(string(data), linter.WithSubjectLineLength(subjectLineLength), linter.WithBodyLineLength(bodyLineLength))
	err = cml.Lint()
	if err != nil {
		log.Printf("linting failed: %s", err)
		os.Exit(1)
	}
}

// buildCommitMessage is used to build a commit message
func buildCommitMessage() error {
	cmb, err := builder.NewCommitMessageBuilder(builder.WithBodyLineLength(bodyLineLength), builder.WithSubjectLineLength(subjectLineLength))
	if err != nil {
		return fmt.Errorf("error creating builder: %w", err)
	}
	if verbose {
		log.Print("asking for data")
	}
	err = cmb.Build()
	if err != nil {
		return fmt.Errorf("error creating commit message: %w", err)
	}
	msg := cmb.String()
	if verbose {
		log.Print("commit message created")
	}

	if printCommitMessage {
		log.Println("resulting commit message:")
		for _, s := range strings.Split(msg, " \n") {
			log.Printf("  %s", s)
		}
	}

	if add {
		if verbose {
			log.Print("stage files")
		}
		_, err = Git("add", "--all", ":/")
		if err != nil {
			return fmt.Errorf("could not add all changes: %s", err)
		}
	}
	if verbose {
		log.Print("commit")
	}
	_, err = Git("commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("could not commit: %s", err)
	}
	if push {
		if verbose {
			log.Print("push")
		}
		_, err = Git("push")
		if err != nil {
			return fmt.Errorf("could not push changes: %s", err)
		}
	}
	return nil
}

// init is known
func init() {
	log.SetPrefix("[git-c] ")
	log.SetFlags(log.LstdFlags | log.LUTC)

	var err error
	gitExecutable, err = exec.LookPath("git")
	if err != nil {
		log.Fatalf("could not locate git: '%#v'", err)
	}
	flag.SetEnvPrefix("GIT_C")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&verbose, "verbose", false, "print more information on execution")
	flag.BoolVar(&add, "add", false, "add all changed files before committing")
	flag.BoolVar(&push, "push", false, "automatically push to default remote")
	flag.BoolVar(&printCommitMessage, "print", false, "print generated commit message")
	flag.StringVar(&commitMessageFile, "lint", "", "print generated commit message")
	flag.IntVar(&subjectLineLength, "subject-line-length", 50, "max length of subject line")
	flag.IntVar(&bodyLineLength, "body-line-length", 72, "max length of a body line")
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
