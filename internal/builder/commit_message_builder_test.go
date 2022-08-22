package builder

import (
	"github.com/sascha-andres/gitc/internal/linter"
	"testing"
)

// builderTestCases contains a list of test cases for the builder
var builderTestCases = map[string]struct {
	cmb CommitMessageBuilder
	err error
}{
	"simple subject line": {
		cmb: CommitMessageBuilder{
			Type:    "feat",
			Message: "abc",
		},
		err: nil,
	},
	"simple subject line with issue": {
		cmb: CommitMessageBuilder{
			Type:    "feat",
			Message: "abc",
			Issue:   "#1",
		},
		err: nil,
	},
	"body": {
		cmb: CommitMessageBuilder{
			Type:    "feat",
			Message: "abc",
			Body:    "abc",
		},
		err: nil,
	},
	"co-authored-by and body": {
		cmb: CommitMessageBuilder{
			Type:      "feat",
			Message:   "abc",
			Body:      "abc",
			CoAuthors: "Co-authored-by: name <email@test.de>",
		},
		err: nil,
	},
}

// TestBuilder iterates over builderTestCases
func TestBuilder(t *testing.T) {
	for s, testCase := range builderTestCases {
		t.Run(s, func(t *testing.T) {
			cml, _ := linter.NewCommitMessageLinter(testCase.cmb.String())
			err := cml.Lint()
			if err != testCase.err {
				t.Fail()
			}
		})
	}
}

// TestSubjectOnly is a testcase for whitespace suffix
func TestSubjectOnly(t *testing.T) {
	cml, _ := linter.NewCommitMessageLinter(`feat: bla
`)
	err := cml.Lint()
	if err != nil {
		t.Errorf("%s", err)
		t.Fail()
	}
}
