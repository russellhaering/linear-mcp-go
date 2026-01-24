package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// AddLabelTool is the tool definition for adding labels to an issue
var AddLabelTool = mcp.NewTool("linear_add_label",
	mcp.WithDescription("Adds one or more labels to an existing Linear issue."),
	mcp.WithString("issue", mcp.Required(), mcp.Description("Issue ID or identifier (e.g., 'TEAM-123')")),
	mcp.WithString("labels", mcp.Required(), mcp.Description("Comma-separated list of label IDs or names to add")),
)

// AddLabelHandler handles the linear_add_label tool
func AddLabelHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		// Get the issue to retrieve its team ID (needed for label resolution)
		issue, err := linearClient.GetIssue(issueID)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get issue: %v", err)}}}, nil
		}

		if issue.Team == nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Issue does not have a team"}}}, nil
		}

		teamID := issue.Team.ID

		// Parse comma-separated labels
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
		newLabelIDs, err := resolveLabelIdentifiers(linearClient, teamID, labelIdentifiers)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve labels: %v", err)}}}, nil
		}

		// Get current labels on the issue and merge with new ones
		var existingLabelIDs []string
		if issue.Labels != nil {
			for _, label := range issue.Labels.Nodes {
				existingLabelIDs = append(existingLabelIDs, label.ID)
			}
		}

		// Create a set of all label IDs (existing + new) to avoid duplicates
		labelIDSet := make(map[string]bool)
		for _, id := range existingLabelIDs {
			labelIDSet[id] = true
		}
		for _, id := range newLabelIDs {
			labelIDSet[id] = true
		}

		// Convert set back to slice
		var allLabelIDs []string
		for id := range labelIDSet {
			allLabelIDs = append(allLabelIDs, id)
		}

		// Update the issue with the combined labels
		updatedIssue, err := linearClient.AddLabelToIssue(issueID, allLabelIDs)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to add labels: %v", err)}}}, nil
		}

		// Build result text
		resultText := fmt.Sprintf("Added label(s) to %s", formatIssueIdentifier(updatedIssue))
		resultText += fmt.Sprintf("\nURL: %s", updatedIssue.URL)

		// List the labels on the issue
		if updatedIssue.Labels != nil && len(updatedIssue.Labels.Nodes) > 0 {
			var labelNames []string
			for _, label := range updatedIssue.Labels.Nodes {
				labelNames = append(labelNames, label.Name)
			}
			resultText += fmt.Sprintf("\nLabels: %s", strings.Join(labelNames, ", "))
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText}}}, nil
	}
}
