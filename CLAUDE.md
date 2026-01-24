# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Linear MCP Server is a Model Context Protocol (MCP) server written in Go that enables AI assistants to interact with Linear issue tracking through standardized MCP tools. It uses GraphQL for all Linear API communication.

## Build and Test Commands

```bash
# Build
go build

# Run all tests (uses VCR cassettes, no API key needed)
go test -v ./...

# Run specific test
go test -v -run TestSpecificFunction ./pkg/server

# Re-record test cassettes (requires LINEAR_API_KEY)
go test -v -record=true ./...

# Update golden files
go test -v -golden=true ./...

# Format and lint
go fmt ./...
go vet ./...
```

## Running the Server

```bash
# Read-only mode (default)
LINEAR_API_KEY=your_key ./linear-mcp-go serve

# With write access
LINEAR_API_KEY=your_key ./linear-mcp-go serve --write-access
```

## Architecture

### Directory Structure
- `cmd/` - Cobra CLI commands (root, serve, version, setup)
- `pkg/server/` - MCP server implementation and tool registration
- `pkg/linear/` - GraphQL API client with rate limiting (1400 req/hour)
- `pkg/tools/` - Individual MCP tool implementations (one file per tool)
- `testdata/fixtures/` - VCR cassettes for recorded HTTP interactions
- `testdata/golden/` - Expected test outputs

### Key Patterns

**Tool Registration**: Tools are defined in `pkg/tools/` and registered in `pkg/server/server.go:RegisterTools()`. Write tools are conditionally registered based on the `writeAccess` flag.

**Entity Identifier Resolution**: Tools accept multiple identifier formats (UUID, human-readable like "TEAM-123", names, or slugs) and resolve them to UUIDs for API calls.

**Rate Limiting**: Queue-based adaptive rate limiter in `pkg/linear/rate_limiter.go` manages Linear's 1400 requests/hour limit.

**Test Infrastructure**: Uses go-vcr to record/replay HTTP interactions. Tests can run without an API key using pre-recorded cassettes. Use `-record=true` flag with `LINEAR_API_KEY` set to update cassettes.

### Adding a New Tool

1. Create a new file in `pkg/tools/` (e.g., `my_tool.go`)
2. Define the tool using `mcp.NewTool()` with parameters
3. Create a handler function that returns `mcpserver.ToolHandlerFunc`
4. Register the tool in `pkg/server/server.go:RegisterTools()`
5. Add test cassettes in `testdata/fixtures/` and golden files in `testdata/golden/`

## Version and Release

Version is stored in `pkg/server/server.go` constant `ServerVersion`. Releases are automated via GitHub Actions when tags matching `v*` are pushed.
