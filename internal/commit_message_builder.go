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
	result, err := promptSelect("Type of change", []string{
		"feat",
		"fix",
		"doc",
		"chore",
		"refactor",
		"test",
		"style",
		"perf",
		"other",
	})
	if err != nil {
		return err
	}
	cmb.Type = strings.TrimSpace(result)

	result, err = promptText("Issue (scope)", nil)
	if err != nil {
		return err
	}
	cmb.Issue = result

	result, err = promptText("Message", func(input string) error {
		if len(input) > 50 {
			return errors.New("message must not be longer than 50 characters")
		}
		if len(input) == 0 {
			return errors.New("message must not be provided")
		}
		return nil
	})
	if err != nil {
		return err
	}
	cmb.Message = result

	result = "-"
	for {
		result, err = promptText("Body, empty to end (repeated)", func(input string) error {
			if len(input) <= 72 {
				return nil
			}
			return errors.New("no line must be longer than 72 characters")
		})
		if err != nil {
			return err
		}
		if len(result) == 0 {
			break
		}
		if cmb.Body == "" {
			cmb.Body = result
		} else {
			cmb.Body = fmt.Sprintf("%s\n%s", cmb.Body, result)
		}
	}

	result = "-"
	for {
		result, err = promptText("Co-Authored-By, empty to end (repeated)", func(input string) error {
			if len(input) == 0 {
				return nil
			}
			if !coAuthoredRegex.MatchString(input) {
				return errors.New("please use the format [another-name <another-name@example.com>]")
			}
			return nil
		})
		if err != nil {
			return err
		}
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

// promptSelect runs a select prompt
func promptSelect(label string, items []string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	_, result, err := prompt.Run()
	return result, err
}

// promptText runs a textual prompt
func promptText(label string, val func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}
	if nil != val {
		prompt.Validate = val
	}
	return prompt.Run()
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
