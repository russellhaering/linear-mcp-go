package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// RemoveLabelTool is the tool definition for removing labels from an issue
var RemoveLabelTool = mcp.NewTool("linear_remove_label",
	mcp.WithDescription("Removes one or more labels from an existing Linear issue."),
	mcp.WithString("issue", mcp.Required(), mcp.Description("Issue ID or identifier (e.g., 'TEAM-123')")),
	mcp.WithString("labels", mcp.Required(), mcp.Description("Comma-separated list of label IDs or names to remove")),
)

// RemoveLabelHandler handles the linear_remove_label tool
func RemoveLabelHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		issueIdentifier, err := request.RequireString("issue")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		labelsStr, err := request.RequireString("labels")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		// Resolve issue identifier to a UUID
		issueID, err := resolveIssueIdentifier(linearClient, issueIdentifier)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve issue: %v", err)}}}, nil
		}

		// Get the issue to retrieve its team ID and current labels
		issue, err := linearClient.GetIssue(issueID)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get issue: %v", err)}}}, nil
		}

		if issue.Team == nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Issue does not have a team"}}}, nil
		}

		teamID := issue.Team.ID

		// Parse comma-separated labels to remove
		var labelIdentifiers []string
		for _, label := range strings.Split(labelsStr, ",") {
			trimmedLabel := strings.TrimSpace(label)
			if trimmedLabel != "" {
				labelIdentifiers = append(labelIdentifiers, trimmedLabel)
			}
		}

		if len(labelIdentifiers) == 0 {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No labels provided"}}}, nil
		}

		// Resolve label identifiers to UUIDs
		labelsToRemove, err := resolveLabelIdentifiers(linearClient, teamID, labelIdentifiers)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve labels: %v", err)}}}, nil
		}

		// Create a set of label IDs to remove for quick lookup
		removeSet := make(map[string]bool)
		for _, id := range labelsToRemove {
			removeSet[id] = true
		}

		// Filter out the labels to remove from current labels
		var remainingLabelIDs []string
		if issue.Labels != nil {
			for _, label := range issue.Labels.Nodes {
				if !removeSet[label.ID] {
					remainingLabelIDs = append(remainingLabelIDs, label.ID)
				}
			}
		}

		// Update the issue with the remaining labels
		updatedIssue, err := linearClient.AddLabelToIssue(issueID, remainingLabelIDs)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to remove labels: %v", err)}}}, nil
		}

		// Build result text
		resultText := fmt.Sprintf("Removed label(s) from %s", formatIssueIdentifier(updatedIssue))
		resultText += fmt.Sprintf("\nURL: %s", updatedIssue.URL)

		// List the remaining labels on the issue
		if updatedIssue.Labels != nil && len(updatedIssue.Labels.Nodes) > 0 {
			var labelNames []string
			for _, label := range updatedIssue.Labels.Nodes {
				labelNames = append(labelNames, label.Name)
			}
			resultText += fmt.Sprintf("\nRemaining labels: %s", strings.Join(labelNames, ", "))
		} else {
			resultText += "\nNo labels remaining"
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText}}}, nil
	}
}
