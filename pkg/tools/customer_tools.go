package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/geropl/linear-mcp-go/pkg/linear"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetCustomersTool is the tool definition for getting customers
var GetCustomersTool = mcp.NewTool("linear_get_customers",
	mcp.WithDescription("Gets a list of customers from Linear."),
	mcp.WithNumber("limit", mcp.Description("Maximum number of customers to return (default: 50)")),
)

// GetCustomersHandler handles the linear_get_customers tool
func GetCustomersHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract optional limit
		limit := request.GetInt("limit", 50)

		customers, err := linearClient.GetCustomers(limit)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get customers: %v", err)}}}, nil
		}

		if len(customers) == 0 {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No customers found"}}}, nil
		}

		// Build result text
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("Found %d customers:\n\n", len(customers)))

		for _, customer := range customers {
			resultText.WriteString(fmt.Sprintf("• %s", customer.Name))
			if customer.Tier != nil {
				resultText.WriteString(fmt.Sprintf(" [Tier: %s]", customer.Tier.DisplayName))
			}
			if customer.Status != nil {
				resultText.WriteString(fmt.Sprintf(" [Status: %s]", customer.Status.DisplayName))
			}
			if len(customer.Domains) > 0 {
				resultText.WriteString(fmt.Sprintf(" (%s)", strings.Join(customer.Domains, ", ")))
			}
			resultText.WriteString(fmt.Sprintf("\n  ID: %s\n", customer.ID))
			resultText.WriteString(fmt.Sprintf("  URL: %s\n", customer.URL))
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText.String()}}}, nil
	}
}

// GetCustomerTool is the tool definition for getting a specific customer
var GetCustomerTool = mcp.NewTool("linear_get_customer",
	mcp.WithDescription("Gets a specific customer by ID, name, or domain. Supports UUID, exact name (case-insensitive), domain, or partial name match."),
	mcp.WithString("customer", mcp.Required(), mcp.Description("Customer identifier: UUID, name (exact or partial match, case-insensitive), or domain")),
)

// GetCustomerHandler handles the linear_get_customer tool
func GetCustomerHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		identifier, err := request.RequireString("customer")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		customer, err := linearClient.GetCustomer(identifier)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get customer: %v", err)}}}, nil
		}

		// Build result text
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("Customer: %s\n", customer.Name))
		resultText.WriteString(fmt.Sprintf("ID: %s\n", customer.ID))
		resultText.WriteString(fmt.Sprintf("URL: %s\n", customer.URL))

		if len(customer.Domains) > 0 {
			resultText.WriteString(fmt.Sprintf("Domains: %s\n", strings.Join(customer.Domains, ", ")))
		}

		if customer.Tier != nil {
			resultText.WriteString(fmt.Sprintf("Tier: %s\n", customer.Tier.DisplayName))
		} else {
			resultText.WriteString("Tier: (none)\n")
		}

		if customer.Status != nil {
			resultText.WriteString(fmt.Sprintf("Status: %s\n", customer.Status.DisplayName))
		}

		if customer.Owner != nil {
			resultText.WriteString(fmt.Sprintf("Owner: %s\n", customer.Owner.Name))
		}

		if customer.Revenue != nil {
			resultText.WriteString(fmt.Sprintf("Revenue: $%d\n", *customer.Revenue))
		}

		if customer.Size != nil {
			resultText.WriteString(fmt.Sprintf("Size: %.0f\n", *customer.Size))
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText.String()}}}, nil
	}
}

// GetCustomerTiersTool is the tool definition for getting customer tiers
var GetCustomerTiersTool = mcp.NewTool("linear_get_customer_tiers",
	mcp.WithDescription("Gets all available customer tiers from Linear."),
)

// GetCustomerTiersHandler handles the linear_get_customer_tiers tool
func GetCustomerTiersHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tiers, err := linearClient.GetCustomerTiers()
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to get customer tiers: %v", err)}}}, nil
		}

		if len(tiers) == 0 {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No customer tiers found. You may need to configure customer tiers in Linear settings."}}}, nil
		}

		// Build result text
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("Found %d customer tiers:\n\n", len(tiers)))

		for _, tier := range tiers {
			resultText.WriteString(fmt.Sprintf("• %s", tier.DisplayName))
			if tier.Name != tier.DisplayName {
				resultText.WriteString(fmt.Sprintf(" (%s)", tier.Name))
			}
			resultText.WriteString(fmt.Sprintf("\n  ID: %s\n", tier.ID))
			if tier.Description != "" {
				resultText.WriteString(fmt.Sprintf("  Description: %s\n", tier.Description))
			}
		}

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText.String()}}}, nil
	}
}

// SetCustomerTierTool is the tool definition for setting a customer's tier
var SetCustomerTierTool = mcp.NewTool("linear_set_customer_tier",
	mcp.WithDescription("Sets the tier for a customer in Linear. The customer can be identified by UUID, exact name (case-insensitive), domain, or partial name match."),
	mcp.WithString("customer", mcp.Required(), mcp.Description("Customer identifier: UUID, name (exact or partial match, case-insensitive), or domain")),
	mcp.WithString("tier", mcp.Required(), mcp.Description("Tier ID or name to set")),
)

// SetCustomerTierHandler handles the linear_set_customer_tier tool
func SetCustomerTierHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		customerIdentifier, err := request.RequireString("customer")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		tierIdentifier, err := request.RequireString("tier")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		// Resolve customer identifier to ID
		customer, err := linearClient.GetCustomer(customerIdentifier)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to find customer: %v", err)}}}, nil
		}

		// Resolve tier identifier to ID
		var tierID string
		if isValidUUID(tierIdentifier) {
			tierID = tierIdentifier
		} else {
			tier, err := linearClient.GetCustomerTierByName(tierIdentifier)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to find tier: %v", err)}}}, nil
			}
			tierID = tier.ID
		}

		// Update the customer
		updatedCustomer, err := linearClient.UpdateCustomer(customer.ID, &tierID, nil)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to update customer tier: %v", err)}}}, nil
		}

		// Build result text
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("Updated customer '%s'\n", updatedCustomer.Name))
		if updatedCustomer.Tier != nil {
			resultText.WriteString(fmt.Sprintf("New tier: %s\n", updatedCustomer.Tier.DisplayName))
		}
		resultText.WriteString(fmt.Sprintf("URL: %s\n", updatedCustomer.URL))

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText.String()}}}, nil
	}
}

// SetCustomerStatusTool is the tool definition for setting a customer's status
var SetCustomerStatusTool = mcp.NewTool("linear_set_customer_status",
	mcp.WithDescription("Sets the status for a customer in Linear. The customer can be identified by UUID, exact name (case-insensitive), domain, or partial name match."),
	mcp.WithString("customer", mcp.Required(), mcp.Description("Customer identifier: UUID, name (exact or partial match, case-insensitive), or domain")),
	mcp.WithString("status", mcp.Required(), mcp.Description("Status ID or name to set")),
)

// SetCustomerStatusHandler handles the linear_set_customer_status tool
func SetCustomerStatusHandler(linearClient *linear.LinearClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		customerIdentifier, err := request.RequireString("customer")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		statusIdentifier, err := request.RequireString("status")
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: err.Error()}}}, nil
		}

		// Resolve customer identifier to ID
		customer, err := linearClient.GetCustomer(customerIdentifier)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to find customer: %v", err)}}}, nil
		}

		// Resolve status identifier to ID
		var statusID string
		if isValidUUID(statusIdentifier) {
			statusID = statusIdentifier
		} else {
			status, err := linearClient.GetCustomerStatusByName(statusIdentifier)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to find status: %v", err)}}}, nil
			}
			statusID = status.ID
		}

		// Update the customer
		updatedCustomer, err := linearClient.UpdateCustomer(customer.ID, nil, &statusID)
		if err != nil {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to update customer status: %v", err)}}}, nil
		}

		// Build result text
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("Updated customer '%s'\n", updatedCustomer.Name))
		if updatedCustomer.Status != nil {
			resultText.WriteString(fmt.Sprintf("New status: %s\n", updatedCustomer.Status.DisplayName))
		}
		resultText.WriteString(fmt.Sprintf("URL: %s\n", updatedCustomer.URL))

		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: resultText.String()}}}, nil
	}
}
