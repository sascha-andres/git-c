package builder

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/sascha-andres/gitc/internal"
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
	CoAuthoredByPrefix = "Co-Authored-By"
)

type (
	// CommitMessageBuilder is used to create a conventional git commit message
	CommitMessageBuilder struct {
		// Type of message like feature, bugfix, etc.
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
		// PrefillScopeRegex is a regex that, when set, is used to extract the issue from the current branch
		PrefillScopeRegex string
		// IssuePrefix is prefixed to a detected issue
		IssuePrefix string
		// verbose if true forces the application to print more information
		verbose bool
	}

	// CommitMessageBuilderOption can be used to set options on commit message builder
	CommitMessageBuilderOption func(cmb *CommitMessageBuilder)
)

// WithVerbose turns verbosity on
func WithVerbose() CommitMessageBuilderOption {
	return func(cmb *CommitMessageBuilder) {
		cmb.verbose = true
	}
}

// WithPrefillScopeRegex sets the regex for matching issue based on branch
func WithPrefillScopeRegex(regularExpression string) CommitMessageBuilderOption {
	return func(cmb *CommitMessageBuilder) {
		cmb.PrefillScopeRegex = regularExpression
	}
}

// WithIssuePrefix sets a prefix that is applied when an issue is detected
func WithIssuePrefix(prefix string) CommitMessageBuilderOption {
	return func(cmb *CommitMessageBuilder) {
		cmb.IssuePrefix = prefix
	}
}

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

	scopeDefault := ""
	if cmb.PrefillScopeRegex != "" {
		r, err := regexp.Compile(cmb.PrefillScopeRegex)
		if err != nil {
			return fmt.Errorf("could not parse regex for scope prefill: %w", err)
		}

		var branchBuffer bytes.Buffer
		_, err = internal.Git(&branchBuffer, "branch", "--show-current")
		if cmb.verbose {
			log.Printf("found branch: %q", strings.TrimSpace(branchBuffer.String()))
		}
		if err != nil {
			return fmt.Errorf("could not read current branch: %w", err)
		}
		branchText := branchBuffer.Bytes()
		if r.Match(branchText) {
			match := r.FindSubmatch(branchText)
			paramsMap := make(map[string]string)
			for i, name := range r.SubexpNames() {
				if i > 0 && i <= len(match) {
					paramsMap[name] = string(match[i])
				}
			}
			if val, ok := paramsMap["scope"]; ok {
				scopeDefault = fmt.Sprintf("%s%s", cmb.IssuePrefix, val)
			} else {
				if cmb.verbose {
					log.Print("regular expression does not match")
				}
			}
		}
	}

	result, err = promptText("Issue (scope)", scopeDefault, nil)
	if err != nil {
		return err
	}
	cmb.Issue = result

	maxLineLength := 50
	if cmb.BodyLineLength != 0 {
		maxLineLength = cmb.SubjectLineLength
	}
	result, err = promptText("Message", "", func(input string) error {
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
		result, err = promptText("Body, empty to end (repeated)", "", func(input string) error {
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
		result, err = promptText("Co-Authored-By, empty to end (repeated)", "", func(input string) error {
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
func promptText(label, defaultValue string, val func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
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
