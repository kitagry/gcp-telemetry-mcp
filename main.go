package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kitagry/gcp-telemetry-mcp/logging"
	"github.com/kitagry/gcp-telemetry-mcp/monitoring"
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

	// Create Cloud Monitoring client
	monitoringClient, err := monitoring.NewCloudMonitoringClient(projectID)
	if err != nil {
		fmt.Printf("Failed to create monitoring client: %v\n", err)
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

	// Add create_metric_descriptor tool
	createMetricTool := mcp.NewTool("create_metric_descriptor",
		mcp.WithDescription("Create a custom metric descriptor in Cloud Monitoring"),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Metric type (e.g., 'custom.googleapis.com/my_metric')"),
		),
		mcp.WithString("metric_kind",
			mcp.Required(),
			mcp.Description("Metric kind: GAUGE, DELTA, or CUMULATIVE"),
		),
		mcp.WithString("value_type",
			mcp.Required(),
			mcp.Description("Value type: BOOL, INT64, DOUBLE, STRING, or DISTRIBUTION"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Description of the metric"),
		),
		mcp.WithString("display_name",
			mcp.Description("Display name for the metric"),
		),
	)

	// Add write_time_series tool
	writeTimeSeresTool := mcp.NewTool("write_time_series",
		mcp.WithDescription("Write time series data to Cloud Monitoring"),
		mcp.WithString("metric_type",
			mcp.Required(),
			mcp.Description("Metric type to write data for"),
		),
		mcp.WithString("resource_type",
			mcp.Required(),
			mcp.Description("Resource type (e.g., 'global', 'gce_instance')"),
		),
		mcp.WithNumber("value",
			mcp.Required(),
			mcp.Description("Metric value to write"),
		),
		mcp.WithObject("metric_labels",
			mcp.Description("Optional metric labels"),
		),
		mcp.WithString("timestamp",
			mcp.Description("Timestamp for the data point (ISO 8601 format, defaults to now)"),
		),
	)

	// Add list_time_series tool
	listTimeSeresTool := mcp.NewTool("list_time_series",
		mcp.WithDescription("List time series data from Cloud Monitoring"),
		mcp.WithString("filter",
			mcp.Required(),
			mcp.Description("Monitoring filter expression (e.g., 'metric.type=\"compute.googleapis.com/instance/cpu/usage\"')"),
		),
		mcp.WithString("start_time",
			mcp.Required(),
			mcp.Description("Start time for the query (ISO 8601 format)"),
		),
		mcp.WithString("end_time",
			mcp.Required(),
			mcp.Description("End time for the query (ISO 8601 format)"),
		),
		mcp.WithObject("aggregation",
			mcp.Description("Optional aggregation configuration"),
		),
	)

	// Add list_metric_descriptors tool
	listMetricDescriptorsTool := mcp.NewTool("list_metric_descriptors",
		mcp.WithDescription("List metric descriptors from Cloud Monitoring"),
		mcp.WithString("filter",
			mcp.Description("Filter expression for metric descriptors"),
		),
	)

	// Add delete_metric_descriptor tool
	deleteMetricTool := mcp.NewTool("delete_metric_descriptor",
		mcp.WithDescription("Delete a custom metric descriptor from Cloud Monitoring"),
		mcp.WithString("metric_type",
			mcp.Required(),
			mcp.Description("Metric type to delete"),
		),
	)

	// Add tool handlers
	s.AddTool(writeLogTool, createWriteLogHandler(loggingClient))
	s.AddTool(listLogsTool, createListLogsHandler(loggingClient))
	s.AddTool(createMetricTool, createMetricDescriptorHandler(monitoringClient))
	s.AddTool(writeTimeSeresTool, createWriteTimeSeriesHandler(monitoringClient))
	s.AddTool(listTimeSeresTool, createListTimeSeriesHandler(monitoringClient))
	s.AddTool(listMetricDescriptorsTool, createListMetricDescriptorsHandler(monitoringClient))
	s.AddTool(deleteMetricTool, createDeleteMetricDescriptorHandler(monitoringClient))

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

// createMetricDescriptorHandler creates a handler for creating metric descriptors
func createMetricDescriptorHandler(client monitoring.MonitoringClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metricType, err := request.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}

		metricKind, err := request.RequireString("metric_kind")
		if err != nil {
			return mcp.NewToolResultError("metric_kind is required"), nil
		}

		valueType, err := request.RequireString("value_type")
		if err != nil {
			return mcp.NewToolResultError("value_type is required"), nil
		}

		description, err := request.RequireString("description")
		if err != nil {
			return mcp.NewToolResultError("description is required"), nil
		}

		args := request.GetArguments()
		displayName := ""
		if displayNameArg, exists := args["display_name"]; exists {
			if dn, ok := displayNameArg.(string); ok {
				displayName = dn
			}
		}

		req := monitoring.CreateMetricRequest{
			MetricDescriptor: monitoring.MetricDescriptor{
				Type:        metricType,
				MetricKind:  metricKind,
				ValueType:   valueType,
				Description: description,
				DisplayName: displayName,
			},
		}

		err = client.CreateMetricDescriptor(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create metric descriptor: %v", err)), nil
		}

		return mcp.NewToolResultText("Metric descriptor created successfully"), nil
	}
}

// createWriteTimeSeriesHandler creates a handler for writing time series data
func createWriteTimeSeriesHandler(client monitoring.MonitoringClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metricType, err := request.RequireString("metric_type")
		if err != nil {
			return mcp.NewToolResultError("metric_type is required"), nil
		}

		resourceType, err := request.RequireString("resource_type")
		if err != nil {
			return mcp.NewToolResultError("resource_type is required"), nil
		}

		valueArg, err := request.RequireFloat("value")
		if err != nil {
			return mcp.NewToolResultError("value is required"), nil
		}

		args := request.GetArguments()

		// Parse timestamp
		timestamp := time.Now()
		if timestampArg, exists := args["timestamp"]; exists {
			if ts, ok := timestampArg.(string); ok && ts != "" {
				if parsedTime, parseErr := time.Parse(time.RFC3339, ts); parseErr == nil {
					timestamp = parsedTime
				}
			}
		}

		// Parse metric labels
		var metricLabels map[string]string
		if labelsArg, exists := args["metric_labels"]; exists {
			if labels, ok := labelsArg.(map[string]any); ok {
				metricLabels = make(map[string]string)
				for k, v := range labels {
					if str, ok := v.(string); ok {
						metricLabels[k] = str
					}
				}
			}
		}

		timeSeries := monitoring.TimeSeriesData{
			MetricType:   metricType,
			MetricLabels: metricLabels,
			ResourceType: resourceType,
			Values: []monitoring.MetricValue{
				{
					Value:     valueArg,
					Timestamp: timestamp,
				},
			},
		}

		req := monitoring.WriteTimeSeriesRequest{
			TimeSeries: []monitoring.TimeSeriesData{timeSeries},
		}

		err = client.WriteTimeSeries(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write time series: %v", err)), nil
		}

		return mcp.NewToolResultText("Time series data written successfully"), nil
	}
}

// createListTimeSeriesHandler creates a handler for listing time series data
func createListTimeSeriesHandler(client monitoring.MonitoringClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filter, err := request.RequireString("filter")
		if err != nil {
			return mcp.NewToolResultError("filter is required"), nil
		}

		startTimeStr, err := request.RequireString("start_time")
		if err != nil {
			return mcp.NewToolResultError("start_time is required"), nil
		}

		endTimeStr, err := request.RequireString("end_time")
		if err != nil {
			return mcp.NewToolResultError("end_time is required"), nil
		}

		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time format: %v", err)), nil
		}

		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid end_time format: %v", err)), nil
		}

		req := monitoring.ListTimeSeriesRequest{
			Filter: filter,
		}
		req.Interval.StartTime = startTime
		req.Interval.EndTime = endTime

		// Parse optional aggregation
		args := request.GetArguments()
		if aggArg, exists := args["aggregation"]; exists {
			if agg, ok := aggArg.(map[string]any); ok {
				aggConfig := &monitoring.AggregationConfig{}

				if alignmentPeriod, exists := agg["alignment_period"]; exists {
					if ap, ok := alignmentPeriod.(string); ok {
						aggConfig.AlignmentPeriod = ap
					}
				}

				if perSeriesAligner, exists := agg["per_series_aligner"]; exists {
					if psa, ok := perSeriesAligner.(string); ok {
						aggConfig.PerSeriesAligner = psa
					}
				}

				if crossSeriesReducer, exists := agg["cross_series_reducer"]; exists {
					if csr, ok := crossSeriesReducer.(string); ok {
						aggConfig.CrossSeriesReducer = csr
					}
				}

				if groupByFields, exists := agg["group_by_fields"]; exists {
					if gbf, ok := groupByFields.([]any); ok {
						for _, field := range gbf {
							if fieldStr, ok := field.(string); ok {
								aggConfig.GroupByFields = append(aggConfig.GroupByFields, fieldStr)
							}
						}
					}
				}

				req.Aggregation = aggConfig
			}
		}

		timeSeries, err := client.ListTimeSeries(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list time series: %v", err)), nil
		}

		// Convert time series to JSON for response
		timeSeriesJSON, err := json.MarshalIndent(timeSeries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal time series: %v", err)), nil
		}

		return mcp.NewToolResultText(string(timeSeriesJSON)), nil
	}
}

// createListMetricDescriptorsHandler creates a handler for listing metric descriptors
func createListMetricDescriptorsHandler(client monitoring.MonitoringClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		filter := ""
		if filterArg, exists := args["filter"]; exists {
			if f, ok := filterArg.(string); ok {
				filter = f
			}
		}

		descriptors, err := client.ListMetricDescriptors(ctx, filter)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list metric descriptors: %v", err)), nil
		}

		// Convert descriptors to JSON for response
		descriptorsJSON, err := json.MarshalIndent(descriptors, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal metric descriptors: %v", err)), nil
		}

		return mcp.NewToolResultText(string(descriptorsJSON)), nil
	}
}

// createDeleteMetricDescriptorHandler creates a handler for deleting metric descriptors
func createDeleteMetricDescriptorHandler(client monitoring.MonitoringClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metricType, err := request.RequireString("metric_type")
		if err != nil {
			return mcp.NewToolResultError("metric_type is required"), nil
		}

		err = client.DeleteMetricDescriptor(ctx, metricType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete metric descriptor: %v", err)), nil
		}

		return mcp.NewToolResultText("Metric descriptor deleted successfully"), nil
	}
}
