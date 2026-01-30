package attachments

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/jira-ticket-cli/api"
	"github.com/open-cli-collective/jira-ticket-cli/internal/cmd/root"
)

// Register registers the attachments commands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:     "attachments",
		Aliases: []string{"attachment", "att"},
		Short:   "Manage issue attachments",
		Long:    "Commands for listing, adding, downloading, and deleting issue attachments.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newAddCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newDeleteCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <issue-key>",
		Aliases: []string{"ls"},
		Short:   "List attachments on an issue",
		Long:    "List all attachments on a Jira issue.",
		Example: `  # List attachments
  jtk attachments list PROJ-123

  # Output as JSON
  jtk attachments list PROJ-123 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts, args[0])
		},
	}

	return cmd
}

func runList(opts *root.Options, issueKey string) error {
	v := opts.View()

	client, err := opts.APIClient()
	if err != nil {
		return err
	}

	attachments, err := client.GetIssueAttachments(issueKey)
	if err != nil {
		return err
	}

	if len(attachments) == 0 {
		v.Info("No attachments found on %s", issueKey)
		return nil
	}

	if opts.Output == "json" {
		return v.JSON(attachments)
	}

	headers := []string{"ID", "FILENAME", "SIZE", "AUTHOR", "CREATED"}
	var rows [][]string

	for _, att := range attachments {
		created := formatDate(att.Created)
		author := att.Author.DisplayName
		if author == "" {
			author = att.Author.AccountID
		}

		rows = append(rows, []string{
			att.ID.String(),
			att.Filename,
			api.FormatFileSize(att.Size),
			author,
			created,
		})
	}

	return v.Table(headers, rows)
}

func newAddCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <issue-key> --file <path>",
		Short: "Add an attachment to an issue",
		Long: `Upload a file as an attachment to a Jira issue.

Multiple files can be attached by repeating the --file flag.`,
		Example: `  # Add a single file
  jtk attachments add PROJ-123 --file document.pdf

  # Add multiple files
  jtk attachments add PROJ-123 --file doc.pdf --file screenshot.png`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			files, _ := cmd.Flags().GetStringArray("file")
			return runAdd(opts, args[0], files)
		},
	}

	cmd.Flags().StringArrayP("file", "f", nil, "File(s) to attach (can be repeated)")
	cmd.MarkFlagRequired("file") //nolint:errcheck

	return cmd
}

func runAdd(opts *root.Options, issueKey string, files []string) error {
	v := opts.View()

	if len(files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	client, err := opts.APIClient()
	if err != nil {
		return err
	}

	var allAttachments []api.Attachment
	for _, file := range files {
		v.Info("Uploading %s...", filepath.Base(file))

		attachments, err := client.AddAttachment(issueKey, file)
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", file, err)
		}

		allAttachments = append(allAttachments, attachments...)
	}

	if opts.Output == "json" {
		return v.JSON(allAttachments)
	}

	for _, att := range allAttachments {
		v.Success("Added %s (ID: %s, %s)", att.Filename, att.ID.String(), api.FormatFileSize(att.Size))
	}

	return nil
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "get <attachment-id>",
		Short: "Download an attachment",
		Long: `Download an attachment by its ID.

The attachment ID can be found using 'jtk attachments list'.`,
		Example: `  # Download to current directory
  jtk attachments get 12345

  # Download to specific directory
  jtk attachments get 12345 --output ./downloads/

  # Download with specific filename
  jtk attachments get 12345 --output ./myfile.pdf`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(opts, args[0], outputPath)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "O", ".", "Output path (directory or file)")

	return cmd
}

func runGet(opts *root.Options, attachmentID, outputPath string) error {
	v := opts.View()

	client, err := opts.APIClient()
	if err != nil {
		return err
	}

	// Get attachment metadata first
	attachment, err := client.GetAttachment(attachmentID)
	if err != nil {
		return err
	}

	v.Info("Downloading %s (%s)...", attachment.Filename, api.FormatFileSize(attachment.Size))

	if err := client.DownloadAttachment(attachment, outputPath); err != nil {
		return err
	}

	// Determine final output path for success message
	finalPath := outputPath
	if isDirectory(outputPath) {
		finalPath = filepath.Join(outputPath, attachment.Filename)
	}

	v.Success("Downloaded to %s", finalPath)
	return nil
}

func newDeleteCmd(opts *root.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <attachment-id>",
		Short: "Delete an attachment",
		Long: `Delete an attachment by its ID.

The attachment ID can be found using 'jtk attachments list'.`,
		Example: `  # Delete an attachment
  jtk attachments delete 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}

	return cmd
}

func runDelete(opts *root.Options, attachmentID string) error {
	v := opts.View()

	client, err := opts.APIClient()
	if err != nil {
		return err
	}

	if err := client.DeleteAttachment(attachmentID); err != nil {
		return err
	}

	v.Success("Deleted attachment %s", attachmentID)
	return nil
}

// formatDate extracts just the date portion from an ISO timestamp
func formatDate(timestamp string) string {
	if len(timestamp) >= 10 {
		return timestamp[:10]
	}
	return timestamp
}

// isDirectory checks if a path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
