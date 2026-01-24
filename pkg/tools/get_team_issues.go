package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetTeamIssuesTool is the tool definition for getting team issues
var GetTeamIssuesTool = mcp.NewTool("linear_get_team_issues",
	mcp.WithDescription("Retrieves issues assigned to a team (excludes issues from child teams)."),
	mcp.WithString("team", mcp.Description("Team identifier (UUID, name, or key)"), mcp.Required()),
	mcp.WithString("status", mcp.Description("Filter by status name (e.g., 'In Progress', 'Done')")),
	mcp.WithString("priority", getPriorityOptions()...),
	mcp.WithString("assignee", mcp.Description("Filter by assignee identifier (UUID, name, or email)")),
	mcp.WithString("labels", mcp.Description("Filter by label names (comma-separated)")),
	mcp.WithBoolean("includeArchived", mcp.Description("Include archived issues in results (default: false)")),
	mcp.WithNumber("limit", mcp.Description("Maximum number of issues to return (default: 50)")),
)

// GetTeamIssuesHandler handles the linear_get_team_issues tool
// Uses the team { issues } query which only returns issues directly assigned to the team,
// excluding issues from child/sub-teams.
func GetTeamIssuesHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Build input for GetTeamIssuesFiltered
		input := linear.GetTeamIssuesInput{}

		// Team is required
		team, err := request.RequireString("team")
		if err != nil || team == "" {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Missing required parameter: team"}}}, nil
		}

		// Resolve team identifier to a team ID
		teamID, err := resolveTeamIdentifier(linearClient, team)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve team: %v", err)}}}, nil
		}
		input.TeamID = teamID

		input.Status = request.GetString("status", "")

		if priorityStr, err := request.RequireString("priority"); err == nil && priorityStr != "" {
			priority, err := parsePriority(priorityStr)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Invalid priority: %v", err)}}}, nil
			}
			input.Priority = &priority
		}

		if assignee, err := request.RequireString("assignee"); err == nil && assignee != "" {
			// Resolve assignee identifier to a user ID
			assigneeID, err := resolveUserIdentifier(linearClient, assignee)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve assignee: %v", err)}}}, nil
			}
			input.AssigneeID = assigneeID
		}

		if labelsStr, err := request.RequireString("labels"); err == nil && labelsStr != "" {
			// Split comma-separated labels
			labels := []string{}
			for _, label := range strings.Split(labelsStr, ",") {
				trimmedLabel := strings.TrimSpace(label)
				if trimmedLabel != "" {
					labels = append(labels, trimmedLabel)
				}
			}
			input.Labels = labels
		}

		input.IncludeArchived = request.GetBool("includeArchived", false)
		input.Limit = request.GetInt("limit", 50)

		// Get team issues using the team { issues } query
		// This excludes child team issues unlike the root issues query with team filter
		issues, err := linearClient.GetTeamIssuesFiltered(input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get team issues: %v", err)}}}, nil
		}

		// Format the result
		resultText := fmt.Sprintf("Found %d issues:\n", len(issues))
		for _, issue := range issues {
			// Create a temporary Issue object to use with formatIssueIdentifier
			tempIssue := &linear.Issue{
				ID:         issue.ID,
				Identifier: issue.Identifier,
			}

			statusStr := "None"
			if issue.Status != "" {
				statusStr = issue.Status
			} else if issue.StateName != "" {
				statusStr = issue.StateName
			}

			resultText += fmt.Sprintf("- %s\n", formatIssueIdentifier(tempIssue))
			resultText += fmt.Sprintf("  Title: %s\n", issue.Title)
			resultText += fmt.Sprintf("  Priority: %s\n", priorityToString(issue.Priority))
			resultText += fmt.Sprintf("  Status: %s\n", statusStr)
			resultText += fmt.Sprintf("  URL: %s\n", issue.URL)
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText}}}, nil
	}
}
