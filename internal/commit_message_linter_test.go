package internal

import "testing"

// lintTestCases contains a list of test cases
var lintTestCases = map[string]struct {
	message string
	err     error
}{
	"empty": {
		message: "",
		err:     ErrNoMessage,
	},
	"subject line without issue": {
		message: "feat: test",
		err:     nil,
	},
	"subject line with issue": {
		message: "feat(#1): test",
		err:     nil,
	},
	"subject line with wrong type": {
		message: "abc: test",
		err:     ErrSubjectLineFormatWrong,
	},
	"empty subject line": {
		message: "\n\nbody",
		err:     ErrEmptySubjectLine,
	},
	"subject line too long": {
		message: "feat: 012345678901234567890123456789012345678901234567891",
		err:     ErrSubjectLineTooLong,
	},
	"missing body": {
		message: "feat: a\n",
		err:     ErrMissingBody,
	},
	"body line too long (one liner)": {
		message: "feat: a\n\n01234567890123456789012345678901234567890123456789012345678901234567890123456789",
		err:     ErrBodyLineTooLong,
	},
	"body line too long (two liner)": {
		message: "feat: a\n\n0123456789012345678901234567890123456789012345678901234567890123456789\n01234567890123456789012345678901234567890123456789012345678901234567890123456789",
		err:     ErrBodyLineTooLong,
	},
	"co-authored-by wrong after body": {
		message: "feat: a\n\n0123456789\n\n01234567890123456789012345678901234567890123456789012345678901234567890123456789",
		err:     ErrCoAuthoredFormatWrong,
	},
	"co-authored-by after body": {
		message: "feat: a\n\n0123456789\n\nCo-authored-by: name <mail@test.de>",
		err:     nil,
	},
	"co-authored-by without body": {
		message: "feat: a\n\nCo-authored-by: name <mail@test.de>",
		err:     nil,
	},
	"co-authored format wrong": {
		message: "feat: a\n\nCo-authored-by: name <mail@test.de>\nabcd",
		err:     ErrCoAuthoredFormatWrong,
	},
	"multiple co-authors": {
		message: "feat: a\n\nCo-authored-by: name <mail@test.de>\nCo-authored-by: name 2 <second@test.de>",
		err:     nil,
	},
	"no line after co-authored": {
		message: "feat: a\n\nCo-authored-by: name <mail@test.de>\n\nabcd",
		err:     ErrNoLineAfterCoAuthored,
	},
	"wrong co-authored-format": {
		message: "feat: abc\n\nabc\n\nname <email@test.de>",
		err:     ErrCoAuthoredFormatWrong,
	},
}

// TestLint tests the linting process
func TestLint(t *testing.T) {
	for key, testCase := range lintTestCases {
		t.Run(key, func(t *testing.T) {
			cml, _ := NewCommitMessageLinter(testCase.message)
			err := cml.Lint()
			if err != testCase.err {
				t.Fail()
			}
		})
	}
}
