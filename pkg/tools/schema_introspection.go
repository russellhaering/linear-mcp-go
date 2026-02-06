package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// IntrospectSchemaTool is the tool definition for introspecting the Linear GraphQL schema
var IntrospectSchemaTool = mcp.NewTool("linear_introspect_schema",
	mcp.WithDescription("Introspect the Linear GraphQL schema to discover available types, fields, and arguments for building GraphQL queries. Returns full schema information or details about a specific type. Use this tool to understand what data you can query with linear_graphql_query."),
	mcp.WithString("type", mcp.Description("Optional: name of a specific type to introspect (e.g., 'Issue', 'Project'). If not provided, returns the full schema.")),
)

// IntrospectSchemaHandler handles the linear_introspect_schema tool
func IntrospectSchemaHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract optional type parameter
		typeName := request.GetString("type", "")

		var query string
		var variables map[string]interface{}

		if typeName != "" {
			// Introspect a specific type
			query = `
				query IntrospectType($name: String!) {
					__type(name: $name) {
						name
						kind
						description
						fields(includeDeprecated: true) {
							name
							description
							type {
								name
								kind
								ofType {
									name
									kind
									ofType {
										name
										kind
									}
								}
							}
							args {
								name
								description
								type {
									name
									kind
									ofType {
										name
										kind
									}
								}
								defaultValue
							}
							isDeprecated
							deprecationReason
						}
						inputFields {
							name
							description
							type {
								name
								kind
								ofType {
									name
									kind
								}
							}
							defaultValue
						}
						enumValues(includeDeprecated: true) {
							name
							description
							isDeprecated
							deprecationReason
						}
						possibleTypes {
							name
							kind
						}
					}
				}
			`
			variables = map[string]interface{}{
				"name": typeName,
			}
		} else {
			// Return full schema introspection
			query = `
				query IntrospectSchema {
					__schema {
						queryType {
							name
						}
						mutationType {
							name
						}
						subscriptionType {
							name
						}
						types {
							name
							kind
							description
						}
						directives {
							name
							description
							locations
							args {
								name
								description
								type {
									name
									kind
								}
								defaultValue
							}
						}
					}
				}
			`
		}

		// Execute the introspection query
		result, err := linearClient.ExecuteRawQuery(query, variables)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Schema introspection failed: %v", err)}}}, nil
		}

		// Format the result as JSON
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to format result: %v", err)}}}, nil
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(resultJSON)}}}, nil
	}
}
