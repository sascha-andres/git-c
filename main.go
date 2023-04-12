package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/sascha-andres/flag"
	"github.com/sascha-andres/gitc/internal"
	"github.com/sascha-andres/gitc/internal/builder"
	"github.com/sascha-andres/gitc/internal/linter"
)

var (
	help, add, printCommitMessage, push, verbose, patch bool
	subjectLineLength, bodyLineLength                   int
	commitMessageFile, prefillScopeRegex, issuePrefix   string
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
	var opts []builder.CommitMessageBuilderOption
	opts = append(opts, builder.WithBodyLineLength(bodyLineLength))
	opts = append(opts, builder.WithSubjectLineLength(subjectLineLength))
	opts = append(opts, builder.WithIssuePrefix(issuePrefix))
	opts = append(opts, builder.WithPrefillScopeRegex(prefillScopeRegex))
	if verbose {
		opts = append(opts, builder.WithVerbose())
	}
	cmb, err := builder.NewCommitMessageBuilder(opts...)
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
		_, err = internal.Git(os.Stdout, "add", "--all", ":/")
		if err != nil {
			return fmt.Errorf("could not add all changes: %s", err)
		}
	}
	if verbose {
		log.Print("commit")
	}
	if patch {
		_, err = internal.Git(os.Stdout, "commit", "-p", "-m", msg)
	} else {
		_, err = internal.Git(os.Stdout, "commit", "-m", msg)
	}
	if err != nil {
		return fmt.Errorf("could not commit: %s", err)
	}
	if push {
		if verbose {
			log.Print("push")
		}
		_, err = internal.Git(os.Stdout, "push")
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

	flag.SetEnvPrefix("GIT_C")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&verbose, "verbose", false, "print more information on execution")
	flag.BoolVar(&add, "add", false, "add all changed files before committing")
	flag.BoolVar(&push, "push", false, "automatically push to default remote")
	flag.BoolVar(&patch, "patch", false, "use patch mode (git commit -p)")
	flag.BoolVar(&printCommitMessage, "print", false, "print generated commit message")
	flag.StringVar(&commitMessageFile, "lint", "", "print generated commit message")
	flag.IntVar(&subjectLineLength, "subject-line-length", 50, "max length of subject line")
	flag.IntVar(&bodyLineLength, "body-line-length", 72, "max length of a body line")
	flag.StringVar(&prefillScopeRegex, "prefill-scope-regex", "", "try to extract scope from branch name")
	flag.StringVar(&issuePrefix, "issue-prefix", "", "prefix detected scope with this value")
}
