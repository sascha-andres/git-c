package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

var (
	l               = log.New(os.Stdout, "[git-c] ", log.LstdFlags)
	coAuthoredRegex = regexp.MustCompile(".*<[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?>")
	help, add       bool
)

type (
	commitMessageBuilder struct {
		Type      string
		Issue     string
		Message   string
		Body      string
		CoAuthors string
	}
)

func main() {

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	cmb := commitMessageBuilder{}

	promptType := promptui.Select{
		Label: "Type of change",
		Items: []string{
			"feat",
			"fix",
			"doc",
			"style",
			"refactor",
			"perf",
			"test",
			"chore",
			"other",
		},
	}
	_, result, err := promptType.Run()
	if err != nil {
		l.Print(err)
		os.Exit(1)
	}
	cmb.Type = strings.TrimSpace(result)

	prompt := promptui.Prompt{
		Label: "Issue",
	}
	result, err = prompt.Run()
	if err != nil {
		l.Print(err)
		os.Exit(1)
	}
	cmb.Issue = result

	prompt.Label = "Message"
	prompt.Validate = func(input string) error {
		if len(input) > 50 {
			return errors.New("message must not be longer than 50 characters")
		}
		if len(input) == 0 {
			return errors.New("message must not be provided")
		}
		return nil
	}
	result, err = prompt.Run()
	if err != nil {
		l.Print(err)
		os.Exit(1)
	}
	cmb.Message = result

	prompt.Label = "Body, empty to end (repeated)"
	prompt.Validate = func(input string) error {
		if len(input) <= 72 {
			return nil
		}
		return errors.New("no line must be longer than 72 characters")
	}
	result = "-"
	for {
		result, err = prompt.Run()
		if len(result) == 0 {
			break
		}
		if cmb.Body == "" {
			cmb.Body = result
		} else {
			cmb.Body = fmt.Sprintf("%s\n%s", cmb.Body, result)
		}
	}

	prompt.Label = "Co-Authored-By, empty to end (repeated)"
	prompt.Validate = func(input string) error {
		if len(input) == 0 {
			return nil
		}
		if !coAuthoredRegex.MatchString(input) {
			return errors.New("please use the format [another-name <another-name@example.com>]")
		}
		return nil
	}
	result = "-"
	for {
		result, err = prompt.Run()
		if len(result) == 0 {
			break
		}
		if cmb.CoAuthors == "" {
			cmb.CoAuthors = result
		} else {
			cmb.CoAuthors = fmt.Sprintf("%s\n%s", cmb.CoAuthors, result)
		}
	}

	if strings.TrimSpace(cmb.Type) == "" {
		l.Print("you need to provide a commit type")
		os.Exit(1)
	}

	if add {
		Git("add", "--all", ":/")
	}
	Git("commit", "-m", cmb.String())
}

func init() {
	var err error
	gitExecutable, err = exec.LookPath("git")
	if err != nil {
		l.Fatalf("could not locate git: '%#v'", err)
	}
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&add, "add", false, "add all changed files before committing")
}

var gitExecutable = ""

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

func (cmb commitMessageBuilder) String() string {
	result := cmb.Type
	if cmb.Issue != "" {
		result = fmt.Sprintf("%s(%s)", result, cmb.Issue)
	}
	result = fmt.Sprintf("%s: %s", result, cmb.Message)
	if cmb.Body != "" {
		result = fmt.Sprintf("%s\n\n%s", result, cmb.Body)
	}
	if cmb.CoAuthors != "" {
		result = fmt.Sprintf("%s\n\n%s", result, cmb.CoAuthors)
	}
	return result
}
