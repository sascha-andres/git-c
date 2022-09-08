# git-c

Loosely inspired by https://github.com/commitizen/cz-cli

## Usage

copy compiled binary as git-c in your path. Instead of calling `git commit ...` call git c

Parameters:

    -help: show usage
    -add: add all unstaged files before committing
    -print: print resulting commit message
    -lint: path to a commit-message
    -verbose: print more information
    -subject-line-length: characters allowed for subject line
    -body-line-length: characters allowed for a body line
    -prefill-scope-regex: provide regex to parse scope from branch
    -issue-prefix: prefix detected scope with this value

### Prefill scope regex

Expected to have a named group scope. Example:

    ^(?P<scope>[0-9]+)

Another example that fits most Jira projects:

    ^(?P<scope>[a-zA-Z0-9]+-[0-9]+)

### Environment variables

Each parameter can be provided using environment variables. Set an environment variable using the prefix `GIT_C_` and the uppercase flag name. That is for help: `GIT_C_HELP`

For boolean flags the value must be `true`
