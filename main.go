package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kitagry/gcp-telemetry-mcp/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Get project ID from environment variable
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Printf("GOOGLE_CLOUD_PROJECT environment variable not set\n")
		os.Exit(1)
	}

	// Create Cloud Logging client
	loggingClient, err := logging.NewCloudLoggingClient(projectID)
	if err != nil {
		fmt.Printf("Failed to create logging client: %v\n", err)
		os.Exit(1)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"GCP Telemetry MCP",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Add write_log_entry tool
	writeLogTool := mcp.NewTool("write_log_entry",
		mcp.WithDescription("Write a log entry to Cloud Logging"),
		mcp.WithString("log_name",
			mcp.Required(),
			mcp.Description("Name of the log to write to"),
		),
		mcp.WithString("severity",
			mcp.Required(),
			mcp.Description("Log severity: DEBUG, INFO, WARNING, ERROR, CRITICAL"),
		),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("Log message"),
		),
		mcp.WithObject("labels",
			mcp.Description("Optional labels for the log entry"),
		),
		mcp.WithObject("payload",
			mcp.Description("Optional structured payload for the log entry"),
		),
	)

	// Add list_log_entries tool
	listLogsTool := mcp.NewTool("list_log_entries",
		mcp.WithDescription("List log entries from Cloud Logging"),
		mcp.WithString("filter",
			mcp.Description("Cloud Logging filter expression"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of entries to return (default: 50)"),
		),
	)

	// Add tool handlers
	s.AddTool(writeLogTool, createWriteLogHandler(loggingClient))
	s.AddTool(listLogsTool, createListLogsHandler(loggingClient))

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// createWriteLogHandler creates a handler for writing log entries
func createWriteLogHandler(client logging.LoggingClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logName, err := request.RequireString("log_name")
		if err != nil {
			return mcp.NewToolResultError("log_name is required"), nil
		}

		severity, err := request.RequireString("severity")
		if err != nil {
			return mcp.NewToolResultError("severity is required"), nil
		}

		message, err := request.RequireString("message")
		if err != nil {
			return mcp.NewToolResultError("message is required"), nil
		}

		// Optional parameters - simplified for now
		var labels map[string]string
		var payload map[string]any

		entry := logging.LogEntry{
			Severity: severity,
			Message:  message,
			Labels:   labels,
			Payload:  payload,
		}

		err = client.WriteEntry(ctx, logName, entry)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write log entry: %v", err)), nil
		}

		return mcp.NewToolResultText("Log entry written successfully"), nil
	}
}

// createListLogsHandler creates a handler for listing log entries
func createListLogsHandler(client logging.LoggingClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		req := logging.ListEntriesRequest{
			Limit: 50, // default
		}

		// Parse optional filter parameter
		args := request.GetArguments()
		if filterArg, exists := args["filter"]; exists {
			if filter, ok := filterArg.(string); ok && filter != "" {
				req.Filter = filter
			}
		}

		// Parse optional limit parameter
		if limitArg, exists := args["limit"]; exists {
			if limitFloat, ok := limitArg.(float64); ok && limitFloat > 0 {
				req.Limit = int(limitFloat)
			}
		}

		entries, err := client.ListEntries(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list log entries: %v", err)), nil
		}

		// Convert entries to JSON for response
		entriesJSON, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal entries: %v", err)), nil
		}

		return mcp.NewToolResultText(string(entriesJSON)), nil
	}
}
