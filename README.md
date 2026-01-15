# jira-ticket-cli

A command-line interface for managing Jira Cloud tickets.

## Installation

### From Source

```bash
git clone https://github.com/piekstra/jira-ticket-cli.git
cd jira-ticket-cli
make install
```

### From Release

Download the latest release from the [releases page](https://github.com/piekstra/jira-ticket-cli/releases).

## Configuration

### Set Credentials

```bash
jira-ticket-cli config set \
  --domain mycompany \
  --email user@example.com \
  --token YOUR_API_TOKEN
```

Get your API token from: https://id.atlassian.com/manage-profile/security/api-tokens

### Environment Variables

Alternatively, set environment variables:

```bash
export JIRA_DOMAIN=mycompany
export JIRA_EMAIL=user@example.com
export JIRA_API_TOKEN=your-token
```

## Usage

### Issues

```bash
# List issues in a project
jira-ticket-cli issues list --project MYPROJECT

# List issues in current sprint
jira-ticket-cli issues list --project MYPROJECT --sprint current

# Get issue details
jira-ticket-cli issues get PROJ-123

# Create an issue
jira-ticket-cli issues create --project MYPROJECT --type Task --summary "Fix login bug"

# Update an issue
jira-ticket-cli issues update PROJ-123 --summary "New summary"
jira-ticket-cli issues update PROJ-123 --field priority=High

# Search with JQL
jira-ticket-cli issues search --jql "project = MYPROJECT AND status = 'In Progress'"

# Assign an issue
jira-ticket-cli issues assign PROJ-123 5b10ac8d82e05b22cc7d4ef5

# List available fields
jira-ticket-cli issues fields
jira-ticket-cli issues fields PROJ-123  # editable fields for specific issue
```

### Transitions

```bash
# List available transitions
jira-ticket-cli transitions list PROJ-123

# Perform a transition
jira-ticket-cli transitions do PROJ-123 "In Progress"
```

### Comments

```bash
# List comments
jira-ticket-cli comments list PROJ-123

# Add a comment
jira-ticket-cli comments add PROJ-123 --body "This is my comment"
```

### Sprints

```bash
# List sprints for a board
jira-ticket-cli sprints list --board 123

# Show current sprint
jira-ticket-cli sprints current --board 123

# List issues in a sprint
jira-ticket-cli sprints issues 456
```

### Boards

```bash
# List boards
jira-ticket-cli boards list
jira-ticket-cli boards list --project MYPROJECT

# Get board details
jira-ticket-cli boards get 123
```

## Output Formats

Use `--output` or `-o` to change output format:

- `table` (default): Human-readable table format
- `json`: JSON format for scripting
- `plain`: Tab-separated values

```bash
jira-ticket-cli issues list --project MYPROJECT -o json
```

## Shell Completion

```bash
# Bash
source <(jira-ticket-cli completion bash)

# Zsh
jira-ticket-cli completion zsh > "${fpath[1]}/_jira-ticket-cli"

# Fish
jira-ticket-cli completion fish | source
```

## Development

```bash
# Build
make build

# Run tests
make test

# Run linter
make lint

# Install locally
make install
```

## License

MIT
