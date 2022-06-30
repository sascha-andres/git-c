package internal

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"regexp"
	"strings"
)

var coAuthoredRegex = regexp.MustCompile(".*<[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?>")

type (
	// CommitMessageBuilder is used to create a conventional git commit message
	CommitMessageBuilder struct {
		// Type of a message like feature, bugfix, etc
		Type string
		// Issue if any
		Issue string
		// Message is the payload of the first line
		Message string
		// Body is a deeper description
		Body string
		// CoAuthors feature a list of other developers who contributed to the code
		CoAuthors string
	}
)

// Build queries for information to create a commit message
func (cmb *CommitMessageBuilder) Build() error {
	promptType := promptui.Select{
		Label: "Type of change",
		Items: []string{
			"feat",
			"fix",
			"doc",
			"chore",
			"refactor",
			"test",
			"style",
			"perf",
			"other",
		},
	}
	_, result, err := promptType.Run()
	if err != nil {
		return err
	}
	cmb.Type = strings.TrimSpace(result)

	prompt := promptui.Prompt{
		Label: "Issue",
	}
	result, err = prompt.Run()
	if err != nil {
		return err
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
		return err
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
			cmb.CoAuthors = fmt.Sprintf("%s\nCo-authored-by: %s", cmb.CoAuthors, result)
		}
	}

	return nil
}

// String implements the stringer interface
func (cmb CommitMessageBuilder) String() string {
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
