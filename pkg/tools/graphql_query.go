package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// GraphQLQueryTool is the tool definition for executing arbitrary read-only GraphQL queries
var GraphQLQueryTool = mcp.NewTool("linear_graphql_query",
	mcp.WithDescription("Execute an arbitrary read-only GraphQL query against the Linear API. Only queries are allowed; mutations and subscriptions are rejected."),
	mcp.WithString("query", mcp.Required(), mcp.Description("The GraphQL query to execute. Must be a query, not a mutation or subscription.")),
	mcp.WithString("variables", mcp.Description("JSON-encoded variables object for the query (optional)")),
)

// mutationOrSubscriptionPattern matches GraphQL mutations and subscriptions at the start of a query
var mutationOrSubscriptionPattern = regexp.MustCompile(`(?i)^\s*(mutation|subscription)\s*[\(\{@a-zA-Z]`)

// GraphQLQueryHandler handles the linear_graphql_query tool
func GraphQLQueryHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract the query parameter
		query, err := request.RequireString("query")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		// Enforce read-only: reject mutations and subscriptions
		trimmedQuery := strings.TrimSpace(query)
		if mutationOrSubscriptionPattern.MatchString(trimmedQuery) {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Only read-only queries are allowed. Mutations and subscriptions are not permitted."}}}, nil
		}

		// Also check if query starts with mutation or subscription keywords (case-insensitive)
		lowerQuery := strings.ToLower(trimmedQuery)
		if strings.HasPrefix(lowerQuery, "mutation") || strings.HasPrefix(lowerQuery, "subscription") {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Only read-only queries are allowed. Mutations and subscriptions are not permitted."}}}, nil
		}

		// Extract optional variables parameter
		var variables map[string]interface{}
		variablesStr := request.GetString("variables", "")
		if variablesStr != "" {
			if err := json.Unmarshal([]byte(variablesStr), &variables); err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to parse variables JSON: %v", err)}}}, nil
			}
		}

		// Execute the query
		result, err := linearClient.ExecuteRawQuery(query, variables)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("GraphQL query failed: %v", err)}}}, nil
		}

		// Format the result as JSON
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to format result: %v", err)}}}, nil
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(resultJSON)}}}, nil
	}
}
