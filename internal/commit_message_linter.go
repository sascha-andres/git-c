package internal

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	ErrNoMessage              = errors.New("no commit message provided")
	ErrEmptySubjectLine       = errors.New("no subject line given")
	ErrSubjectLineFormatWrong = errors.New("subject line does not adhere to required format (type: message or type(issue): message)")
	ErrSubjectLineTooLong     = errors.New("subject line should not be longer than 50 characters (without type and issue)")
	ErrMissingBody            = errors.New("body is missing (subject + empty line)")
	ErrBodyLineTooLong        = errors.New("a body line is too long (max 72 characters)")
	ErrCoAuthoredFormatWrong  = errors.New("co authored line format wrong (Co-authored-by: name <email>)")
	ErrNoLineAfterCoAuthored  = errors.New("no content after co-authored-by lines allowed")

	subjectRegex      = regexp.MustCompile("(?P<type>feat|fix|doc|chore|refactor|test|style|perf|other)(?P<issue>\\([^)]+\\))?: (?P<message>.*)")
	coAuthoredByRegex = regexp.MustCompile("^Co-authored-by: (?P<name>.*) <(?P<mail>[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)>$")
)

type (
	// CommitMessageLinter is used to lint a commit message
	CommitMessageLinter struct {
		// message is the commit message
		message string

		// logger is used to write out some messages
		logger *log.Logger
	}
)

// NewCommitMessageLinter creates a new linter
func NewCommitMessageLinter(msg string) (*CommitMessageLinter, error) {
	return &CommitMessageLinter{message: msg, logger: log.New(os.Stdout, "[cml] ", log.LstdFlags|log.LUTC)}, nil
}

// Lint runs the linter
func (cml *CommitMessageLinter) Lint() error {
	if len(cml.message) == 0 {
		return ErrNoMessage
	}

	lines := strings.Split(cml.message, "\n")

	if err := cml.subjectLine(lines[0]); err != nil {
		return err
	}

	if len(lines) == 1 {
		return nil
	}

	if len(lines) == 2 {
		return ErrMissingBody
	}

	hasToStartWithCoAuthoredBy := false
	for _, s := range lines[2:] {
		if "" == s && hasToStartWithCoAuthoredBy {
			return ErrNoLineAfterCoAuthored
		}
		if hasToStartWithCoAuthoredBy {
			if !coAuthoredByRegex.MatchString(s) {
				return ErrCoAuthoredFormatWrong
			}
		}
		if "" == s && !hasToStartWithCoAuthoredBy {
			hasToStartWithCoAuthoredBy = true
			continue
		}
		if coAuthoredByRegex.MatchString(s) {
			hasToStartWithCoAuthoredBy = true
			continue
		}
		if len(s) > 72 {
			return ErrBodyLineTooLong
		}
	}

	return nil
}

// subjectLine tests the subject line of a commit
func (cml *CommitMessageLinter) subjectLine(line string) error {
	if len(line) == 0 {
		return ErrEmptySubjectLine
	}
	if !subjectRegex.MatchString(line) {
		return ErrSubjectLineFormatWrong
	}
	matches := subjectRegex.FindStringSubmatch(line)
	groupNames := subjectRegex.SubexpNames()
	for i, match := range matches {
		if groupNames[i] == "message" {
			if len(match) > 50 {
				return ErrSubjectLineTooLong
			}
		}
	}
	return nil
}
