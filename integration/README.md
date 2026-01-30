# Integration Tests

These tests run against real Jira APIs and require valid credentials.

## Prerequisites

Set the following environment variables:

```bash
export JIRA_URL="https://your-domain.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="your-api-token"

# For tests that create issues, you need a project key where you have permission
export JIRA_TEST_PROJECT="TEST"

# Optional: specify the issue type to use (default: "Task")
# Some projects may use different names like "SDLC", "Story", "Bug", etc.
export JIRA_TEST_ISSUE_TYPE="Task"
```

Alternatively, use `ATLASSIAN_URL`, `ATLASSIAN_EMAIL`, `ATLASSIAN_API_TOKEN` if you have those configured.

## Running Tests

```bash
# Run all integration tests
go test -tags=integration ./integration/...

# Run specific test file
go test -tags=integration ./integration/... -run TestAttachments

# Run with verbose output
go test -tags=integration -v ./integration/...

# Run a single test
go test -tags=integration -v ./integration/... -run TestAttachments_FullFlow
```

## When to Run

Run integration tests when:

- Modifying API client code in `api/`
- Adding new API functionality
- Changing authentication or request handling
- Debugging production issues
- Before releasing a new version

## Test Behavior

- Tests skip automatically if credentials are not configured
- Tests clean up after themselves (delete created issues, attachments, etc.)
- Tests use unique identifiers to avoid conflicts with parallel runs
- Some tests may be skipped if the Jira instance doesn't support certain features

## Adding New Tests

1. Create a new file with `_test.go` suffix
2. Add the `//go:build integration` build tag at the top
3. Use `skipIfNoCredentials(t)` at the start of each test
4. Clean up any created resources in a `defer` or `t.Cleanup()`

Example:

```go
//go:build integration

package integration

func TestMyFeature(t *testing.T) {
    skipIfNoCredentials(t)
    client := newTestClient(t)

    // Create test data
    // ...

    // Clean up
    t.Cleanup(func() {
        // Delete created resources
    })

    // Test assertions
    // ...
}
```
