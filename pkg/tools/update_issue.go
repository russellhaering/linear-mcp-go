package tools

import (
	"context"
	"fmt"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// UpdateIssueTool is the tool definition for updating issues
var UpdateIssueTool = mcp.NewTool("linear_update_issue",
	mcp.WithDescription("Updates an existing Linear issue."),
	mcp.WithString("issue", mcp.Required(), mcp.Description("Issue ID or identifier (e.g., 'TEAM-123')")),
	mcp.WithString("title", mcp.Description("New title")),
	mcp.WithString("description", mcp.Description("New description")),
	mcp.WithString("priority", getPriorityOptions()...),
	mcp.WithString("status", mcp.Description("New status")),
	mcp.WithString("assignee", mcp.Description("New assignee (UUID, name, or email)")),
	mcp.WithString("team", mcp.Description("New team (UUID, name, or key)")),
	mcp.WithString("projectId", mcp.Description("New project (UUID, name, or slug)")),
	mcp.WithString("milestoneId", mcp.Description("New milestone (UUID or name)")),
)

// UpdateIssueHandler handles the linear_update_issue tool
func UpdateIssueHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		issueIdentifier, err := request.RequireString("issue")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		// Resolve issue identifier to a UUID
		id, err := resolveIssueIdentifier(linearClient, issueIdentifier)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve issue: %v", err)}}}, nil
		}

		// Extract optional arguments
		title := request.GetString("title", "")
		description := request.GetString("description", "")

		var priority *int
		if priorityStr, err := request.RequireString("priority"); err == nil && priorityStr != "" {
			p, err := parsePriority(priorityStr)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Invalid priority: %v", err)}}}, nil
			}
			priority = &p
		}

		// Resolve team identifier to a team ID
		var teamID string
		team := request.GetString("team", "")
		if team != "" {
			// Resolve team identifier to a team ID
			teamID, err = resolveTeamIdentifier(linearClient, team)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve team: %v", err)}}}, nil
			}
		}

		status := request.GetString("status", "")

		// Resolve project identifier to a project ID
		var projectID string
		if projectStr := request.GetString("projectId", ""); projectStr != "" {
			projectID, err = resolveProjectIdentifier(linearClient, projectStr)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve project: %v", err)}}}, nil
			}
		}

		// Resolve milestone identifier to a milestone ID
		var milestoneID string
		if milestoneStr := request.GetString("milestoneId", ""); milestoneStr != "" {
			milestoneID, err = resolveMilestoneIdentifier(linearClient, milestoneStr)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve milestone: %v", err)}}}, nil
			}
		}

		// Resolve status identifier to a state ID
		var stateID string
		if status != "" {
			// To resolve a status name, we need the team ID.
			// Use the provided team ID, or fetch the issue to get it.
			statusTeamID := teamID
			if statusTeamID == "" {
				issue, err := linearClient.GetIssue(id)
				if err != nil {
					return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get issue for status resolution: %v", err)}}}, nil
				}
				if issue.Team != nil {
					statusTeamID = issue.Team.ID
				}
			}
			if statusTeamID == "" {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Cannot resolve status: unable to determine issue team"}}}, nil
			}
			stateID, err = resolveStatusIdentifier(linearClient, statusTeamID, status)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve status: %v", err)}}}, nil
			}
		}

		// Resolve assignee identifier to a user ID
		var assigneeID string
		assignee := request.GetString("assignee", "")
		if assignee != "" {
			assigneeID, err = resolveUserIdentifier(linearClient, assignee)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve assignee: %v", err)}}}, nil
			}
		}

		// Update the issue
		input := linear.UpdateIssueInput{
			ID:          id,
			Title:       title,
			Description: description,
			Priority:    priority,
			Status:      stateID,
			TeamID:      teamID,
			ProjectID:   projectID,
			MilestoneID: milestoneID,
			AssigneeID:  assigneeID,
		}

		issue, err := linearClient.UpdateIssue(input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to update issue: %v", err)}}}, nil
		}

		// Return the result
		resultText := fmt.Sprintf("Updated %s", formatIssueIdentifier(issue))
		resultText += fmt.Sprintf("\nURL: %s", issue.URL)
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText}}}, nil
	}
}
