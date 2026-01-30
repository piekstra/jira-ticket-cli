//go:build integration

package integration

import (
	"testing"

	"github.com/open-cli-collective/jira-ticket-cli/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMoveIssue tests the full move issue flow.
// This requires JIRA_TEST_PROJECT and JIRA_TEST_MOVE_TARGET_PROJECT environment variables.
func TestMoveIssue(t *testing.T) {
	skipIfNoCredentials(t)

	sourceProject := getTestProject(t)
	targetProject := getTestMoveTargetProject(t)
	_ = getTestIssueType(t) // validate test setup

	client := newTestClient(t)

	// Create a test issue in the source project
	issueKey := createTestIssue(t, client, sourceProject, "Integration test - move issue")
	t.Logf("Created test issue: %s", issueKey)

	// Get target project issue types
	issueTypes, err := client.GetProjectIssueTypes(targetProject)
	require.NoError(t, err, "failed to get target project issue types")
	require.NotEmpty(t, issueTypes, "target project has no issue types")

	// Find a non-subtask issue type in target project
	var targetIssueType *api.IssueType
	for i := range issueTypes {
		if !issueTypes[i].Subtask {
			targetIssueType = &issueTypes[i]
			break
		}
	}
	require.NotNil(t, targetIssueType, "no non-subtask issue type found in target project")

	t.Logf("Moving %s to project %s with issue type %s (ID: %s)",
		issueKey, targetProject, targetIssueType.Name, targetIssueType.ID)

	// Build and execute the move request
	req := api.BuildMoveRequest([]string{issueKey}, targetProject, targetIssueType.ID, false)
	resp, err := client.MoveIssues(req)
	require.NoError(t, err, "failed to initiate move")
	assert.NotEmpty(t, resp.TaskID, "task ID should not be empty")

	t.Logf("Move task ID: %s", resp.TaskID)

	// Poll for completion
	var status *api.MoveTaskStatus
	for i := 0; i < 30; i++ { // max 30 seconds
		status, err = client.GetMoveTaskStatus(resp.TaskID)
		require.NoError(t, err, "failed to get task status")

		t.Logf("Task status: %s (progress: %d%%)", status.Status, status.Progress)

		if status.Status == "COMPLETE" || status.Status == "FAILED" || status.Status == "CANCELLED" {
			break
		}

		// Wait a second before polling again
		// Note: In real tests you'd use time.Sleep, but for integration tests
		// we want to be able to observe progress
	}

	require.Equal(t, "COMPLETE", status.Status, "move task should complete successfully")

	if status.Result != nil {
		if len(status.Result.Failed) > 0 {
			t.Errorf("Some issues failed to move: %+v", status.Result.Failed)
		}
		if len(status.Result.Successful) > 0 {
			t.Logf("Successfully moved: %v", status.Result.Successful)
			// The issue key changes after move
			newKey := status.Result.Successful[0]

			// Verify the issue is now in the target project
			issue, err := client.GetIssue(newKey)
			require.NoError(t, err, "failed to get moved issue")
			assert.Equal(t, targetProject, issue.Fields.Project.Key, "issue should be in target project")

			// Clean up - delete the moved issue
			err = client.DeleteIssue(newKey)
			if err != nil {
				t.Logf("Warning: failed to delete test issue %s: %v", newKey, err)
			} else {
				t.Logf("Cleaned up test issue %s", newKey)
			}
		}
	}
}

// TestMoveMultipleIssues tests moving multiple issues at once.
func TestMoveMultipleIssues(t *testing.T) {
	skipIfNoCredentials(t)

	sourceProject := getTestProject(t)
	targetProject := getTestMoveTargetProject(t)

	client := newTestClient(t)

	// Create test issues
	issueKey1 := createTestIssue(t, client, sourceProject, "Integration test - bulk move 1")
	issueKey2 := createTestIssue(t, client, sourceProject, "Integration test - bulk move 2")
	t.Logf("Created test issues: %s, %s", issueKey1, issueKey2)

	// Get target project issue types
	issueTypes, err := client.GetProjectIssueTypes(targetProject)
	require.NoError(t, err)

	var targetIssueType *api.IssueType
	for i := range issueTypes {
		if !issueTypes[i].Subtask {
			targetIssueType = &issueTypes[i]
			break
		}
	}
	require.NotNil(t, targetIssueType)

	// Move both issues
	req := api.BuildMoveRequest([]string{issueKey1, issueKey2}, targetProject, targetIssueType.ID, false)
	resp, err := client.MoveIssues(req)
	require.NoError(t, err)

	// Wait for completion
	var status *api.MoveTaskStatus
	for i := 0; i < 30; i++ {
		status, err = client.GetMoveTaskStatus(resp.TaskID)
		require.NoError(t, err)

		if status.Status == "COMPLETE" || status.Status == "FAILED" || status.Status == "CANCELLED" {
			break
		}
	}

	require.Equal(t, "COMPLETE", status.Status)

	if status.Result != nil {
		assert.Empty(t, status.Result.Failed, "no issues should fail to move")
		assert.Len(t, status.Result.Successful, 2, "both issues should be moved")

		// Clean up
		for _, key := range status.Result.Successful {
			err := client.DeleteIssue(key)
			if err != nil {
				t.Logf("Warning: failed to delete %s: %v", key, err)
			}
		}
	}
}

// TestBuildMoveRequest tests the request building function.
func TestBuildMoveRequest(t *testing.T) {
	req := api.BuildMoveRequest([]string{"PROJ-1", "PROJ-2"}, "TARGET", "10001", true)

	assert.True(t, req.SendBulkNotification)
	assert.Len(t, req.TargetToSourcesMapping, 1)

	// Key format should be "PROJECT,ISSUE_TYPE_ID" (comma-separated)
	spec, exists := req.TargetToSourcesMapping["TARGET,10001"]
	assert.True(t, exists, "target key should use comma separator")
	assert.Equal(t, []string{"PROJ-1", "PROJ-2"}, spec.IssueIdsOrKeys)
	assert.True(t, spec.InferFieldDefaults)
	assert.True(t, spec.InferStatusDefaults)
}
