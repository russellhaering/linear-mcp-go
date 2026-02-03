package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetProjectsTool is the tool definition for getting projects
var GetProjectsTool = mcp.NewTool("linear_get_projects",
	mcp.WithDescription("Get projects, optionally filtered by team."),
	mcp.WithString("team", mcp.Description("Optional team identifier (UUID, name, or key) to filter projects by")),
	mcp.WithNumber("limit", mcp.Description("Maximum number of projects to return (default: 50)")),
)

// GetProjectsHandler handles the linear_get_projects tool
func GetProjectsHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := linear.GetProjectsInput{}

		// Team is optional
		if team := request.GetString("team", ""); team != "" {
			// Resolve team identifier to a team ID
			teamID, err := resolveTeamIdentifier(linearClient, team)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to resolve team: %v", err)}}}, nil
			}
			input.TeamID = teamID
		}

		input.Limit = request.GetInt("limit", 50)

		// Get projects
		projects, err := linearClient.GetProjects(input)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get projects: %v", err)}}}, nil
		}

		if len(projects) == 0 {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No projects found."}}}, nil
		}

		// Format the result
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Found %d projects:\n\n", len(projects)))
		for _, project := range projects {
			builder.WriteString(FormatProject(project))
			builder.WriteString("\n")
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: builder.String()}}}, nil
	}
}
