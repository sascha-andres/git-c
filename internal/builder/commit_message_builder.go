package builder

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"regexp"
	"strings"
)

var (
	coAuthoredRegex = regexp.MustCompile(".*<[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?>")
	types           = []string{
		"feat",
		"fix",
		"doc",
		"chore",
		"refactor",
		"test",
		"style",
		"perf",
		"other",
	}
)

const (
	CoAuthoredByPrefix = "Co-authored-by"
)

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
		// BodyLineLength restricts the length of a body line
		BodyLineLength int
		// SubjectLineLength restricts the length of the subject line
		SubjectLineLength int
	}

	// CommitMessageBuilderOption can be used to set options on commit message builder
	CommitMessageBuilderOption func(cmb *CommitMessageBuilder)
)

// WithBodyLineLength allows setting the maximum length of a body line
func WithBodyLineLength(length int) CommitMessageBuilderOption {
	return func(cmb *CommitMessageBuilder) {
		cmb.BodyLineLength = length
	}
}

// WithSubjectLineLength allows setting the maximum length of a body line
func WithSubjectLineLength(length int) CommitMessageBuilderOption {
	return func(cmb *CommitMessageBuilder) {
		cmb.SubjectLineLength = length
	}
}

// NewCommitMessageBuilder returns a new commit message builder
func NewCommitMessageBuilder(options ...CommitMessageBuilderOption) (*CommitMessageBuilder, error) {
	cmb := &CommitMessageBuilder{}
	for i := range options {
		options[i](cmb)
	}
	return cmb, nil
}

// Build queries for information to create a commit message
func (cmb *CommitMessageBuilder) Build() error {
	result, err := promptSelect("Type of change", types)
	if err != nil {
		return err
	}
	cmb.Type = strings.TrimSpace(result)

	result, err = promptText("Issue (scope)", nil)
	if err != nil {
		return err
	}
	cmb.Issue = result

	maxLineLength := 50
	if cmb.BodyLineLength != 0 {
		maxLineLength = cmb.SubjectLineLength
	}
	result, err = promptText("Message", func(input string) error {
		if len(input) > maxLineLength {
			return errors.New(fmt.Sprintf("message must not be longer than %d characters", maxLineLength))
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
	maxLineLength = 72
	if cmb.BodyLineLength != 0 {
		maxLineLength = cmb.BodyLineLength
	}
	for {
		result, err = promptText("Body, empty to end (repeated)", func(input string) error {
			if len(input) <= maxLineLength {
				return nil
			}
			return errors.New(fmt.Sprintf("no line must be longer than %d characters", maxLineLength))
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
			cmb.CoAuthors = fmt.Sprintf("\n%s: %s", CoAuthoredByPrefix, result)
		} else {
			cmb.CoAuthors = fmt.Sprintf("%s\n%s: %s", cmb.CoAuthors, CoAuthoredByPrefix, result)
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
func (cmb *CommitMessageBuilder) String() string {
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
