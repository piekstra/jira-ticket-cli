//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/open-cli-collective/jira-ticket-cli/api"
	"github.com/open-cli-collective/jira-ticket-cli/internal/config"
)

// skipIfNoCredentials skips the test if Jira credentials are not configured
func skipIfNoCredentials(t *testing.T) {
	t.Helper()

	url := config.GetURL()
	email := config.GetEmail()
	token := config.GetAPIToken()

	if url == "" || email == "" || token == "" {
		t.Skip("Jira credentials not configured (set JIRA_URL, JIRA_EMAIL, JIRA_API_TOKEN or ATLASSIAN_* equivalents)")
	}
}

// getTestProject returns the project key for integration tests
func getTestProject(t *testing.T) string {
	t.Helper()

	project := os.Getenv("JIRA_TEST_PROJECT")
	if project == "" {
		t.Skip("JIRA_TEST_PROJECT not set")
	}
	return project
}

// getTestIssueType returns the issue type to use for integration tests
// Defaults to "Task" but can be overridden with JIRA_TEST_ISSUE_TYPE
func getTestIssueType(t *testing.T) string {
	t.Helper()

	issueType := os.Getenv("JIRA_TEST_ISSUE_TYPE")
	if issueType == "" {
		issueType = "Task"
	}
	return issueType
}

// newTestClient creates an API client for integration tests
func newTestClient(t *testing.T) *api.Client {
	t.Helper()

	client, err := api.New(api.ClientConfig{
		URL:      config.GetURL(),
		Email:    config.GetEmail(),
		APIToken: config.GetAPIToken(),
	})
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	return client
}

// createTestIssue creates a temporary issue for testing and returns its key
// The issue is automatically deleted when the test completes
func createTestIssue(t *testing.T, client *api.Client, project, summary string) string {
	t.Helper()

	issueType := getTestIssueType(t)
	req := api.BuildCreateRequest(project, issueType, summary, "", nil)
	issue, err := client.CreateIssue(req)
	if err != nil {
		t.Fatalf("Failed to create test issue (project=%s, type=%s): %v", project, issueType, err)
	}

	t.Cleanup(func() {
		if err := client.DeleteIssue(issue.Key); err != nil {
			t.Logf("Warning: failed to delete test issue %s: %v", issue.Key, err)
		}
	})

	return issue.Key
}
