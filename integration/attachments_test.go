//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachments_FullFlow(t *testing.T) {
	skipIfNoCredentials(t)
	project := getTestProject(t)
	client := newTestClient(t)

	// Create a test issue to attach files to
	issueKey := createTestIssue(t, client, project, "[Integration Test] Attachment test - safe to delete")

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-attachment.txt")
	testContent := []byte("Integration test attachment content")
	require.NoError(t, os.WriteFile(testFile, testContent, 0644))

	// Test 1: Add attachment
	t.Run("AddAttachment", func(t *testing.T) {
		attachments, err := client.AddAttachment(issueKey, testFile)
		require.NoError(t, err, "AddAttachment should succeed")
		require.Len(t, attachments, 1, "Should return one attachment")

		att := attachments[0]
		assert.Equal(t, "test-attachment.txt", att.Filename)
		assert.Equal(t, int64(len(testContent)), att.Size)
		assert.NotEmpty(t, att.ID.String(), "Attachment ID should not be empty")
	})

	// Test 2: List attachments
	var attachmentID string
	t.Run("ListAttachments", func(t *testing.T) {
		attachments, err := client.GetIssueAttachments(issueKey)
		require.NoError(t, err, "GetIssueAttachments should succeed")
		require.NotEmpty(t, attachments, "Issue should have attachments")

		// Find our attachment
		found := false
		for _, att := range attachments {
			if att.Filename == "test-attachment.txt" {
				found = true
				attachmentID = att.ID.String()
				assert.Equal(t, int64(len(testContent)), att.Size)
				break
			}
		}
		assert.True(t, found, "Should find the uploaded attachment")
	})

	// Test 3: Get attachment metadata
	t.Run("GetAttachment", func(t *testing.T) {
		require.NotEmpty(t, attachmentID, "Need attachment ID from previous test")

		att, err := client.GetAttachment(attachmentID)
		require.NoError(t, err, "GetAttachment should succeed")

		assert.Equal(t, "test-attachment.txt", att.Filename)
		assert.Equal(t, int64(len(testContent)), att.Size)
		assert.NotEmpty(t, att.Content, "Should have content URL")
	})

	// Test 4: Download attachment
	t.Run("DownloadAttachment", func(t *testing.T) {
		require.NotEmpty(t, attachmentID, "Need attachment ID from previous test")

		att, err := client.GetAttachment(attachmentID)
		require.NoError(t, err)

		downloadPath := filepath.Join(tmpDir, "downloaded.txt")
		err = client.DownloadAttachment(att, downloadPath)
		require.NoError(t, err, "DownloadAttachment should succeed")

		// Verify downloaded content
		downloaded, err := os.ReadFile(downloadPath)
		require.NoError(t, err)
		assert.Equal(t, testContent, downloaded, "Downloaded content should match original")
	})

	// Test 5: Download to directory (uses original filename)
	t.Run("DownloadToDirectory", func(t *testing.T) {
		require.NotEmpty(t, attachmentID, "Need attachment ID from previous test")

		att, err := client.GetAttachment(attachmentID)
		require.NoError(t, err)

		downloadDir := filepath.Join(tmpDir, "downloads")
		require.NoError(t, os.MkdirAll(downloadDir, 0755))

		err = client.DownloadAttachment(att, downloadDir)
		require.NoError(t, err, "DownloadAttachment to directory should succeed")

		// Verify file was created with original filename
		downloadedPath := filepath.Join(downloadDir, "test-attachment.txt")
		downloaded, err := os.ReadFile(downloadedPath)
		require.NoError(t, err)
		assert.Equal(t, testContent, downloaded)
	})

	// Test 6: Delete attachment
	t.Run("DeleteAttachment", func(t *testing.T) {
		require.NotEmpty(t, attachmentID, "Need attachment ID from previous test")

		err := client.DeleteAttachment(attachmentID)
		require.NoError(t, err, "DeleteAttachment should succeed")

		// Verify deletion
		attachments, err := client.GetIssueAttachments(issueKey)
		require.NoError(t, err)

		for _, att := range attachments {
			assert.NotEqual(t, attachmentID, att.ID.String(), "Attachment should be deleted")
		}
	})
}

func TestAttachments_MultipleFiles(t *testing.T) {
	skipIfNoCredentials(t)
	project := getTestProject(t)
	client := newTestClient(t)

	issueKey := createTestIssue(t, client, project, "[Integration Test] Multiple attachments - safe to delete")

	tmpDir := t.TempDir()

	// Create multiple test files
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		require.NoError(t, os.WriteFile(path, []byte("Content of "+name), 0644))

		_, err := client.AddAttachment(issueKey, path)
		require.NoError(t, err, "Should add attachment %s", name)
	}

	// Verify all attachments exist
	attachments, err := client.GetIssueAttachments(issueKey)
	require.NoError(t, err)
	assert.Len(t, attachments, len(files), "Should have all attachments")

	// Clean up
	for _, att := range attachments {
		require.NoError(t, client.DeleteAttachment(att.ID.String()))
	}
}

func TestAttachments_ErrorCases(t *testing.T) {
	skipIfNoCredentials(t)
	client := newTestClient(t)

	t.Run("ListAttachments_InvalidIssue", func(t *testing.T) {
		_, err := client.GetIssueAttachments("INVALID-99999")
		assert.Error(t, err, "Should fail for invalid issue")
	})

	t.Run("GetAttachment_InvalidID", func(t *testing.T) {
		_, err := client.GetAttachment("99999999")
		assert.Error(t, err, "Should fail for invalid attachment ID")
	})

	t.Run("DeleteAttachment_InvalidID", func(t *testing.T) {
		err := client.DeleteAttachment("99999999")
		assert.Error(t, err, "Should fail for invalid attachment ID")
	})

	t.Run("AddAttachment_NonexistentFile", func(t *testing.T) {
		project := getTestProject(t)
		issueKey := createTestIssue(t, client, project, "[Integration Test] Error case - safe to delete")

		_, err := client.AddAttachment(issueKey, "/nonexistent/file.txt")
		assert.Error(t, err, "Should fail for nonexistent file")
	})
}
