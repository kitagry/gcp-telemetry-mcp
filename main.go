package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kitagry/gcp-telemetry-mcp/logging"
	"github.com/kitagry/gcp-telemetry-mcp/monitoring"
	"github.com/kitagry/gcp-telemetry-mcp/profiler"
	"github.com/kitagry/gcp-telemetry-mcp/trace"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gcp-telemetry-mcp %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	// Get project ID from environment variable
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Printf("GOOGLE_CLOUD_PROJECT environment variable not set\n")
		os.Exit(1)
	}

	// Create Cloud Logging client
	loggingClient, err := logging.New(projectID)
	if err != nil {
		fmt.Printf("Failed to create logging client: %v\n", err)
		os.Exit(1)
	}

	// Create Cloud Monitoring client
	monitoringClient, err := monitoring.New(projectID)
	if err != nil {
		fmt.Printf("Failed to create monitoring client: %v\n", err)
		os.Exit(1)
	}

	// Create Cloud Trace client
	traceClient, err := trace.New(projectID)
	if err != nil {
		fmt.Printf("Failed to create trace client: %v\n", err)
		os.Exit(1)
	}

	// Create Cloud Profiler client
	profilerClient, err := profiler.New(projectID)
	if err != nil {
		fmt.Printf("Failed to create profiler client: %v\n", err)
		os.Exit(1)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"GCP Telemetry MCP",
		version,
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
			mcp.Description(`Filter expression for metric descriptors.
If this field is empty, all custom and system-defined metric descriptors are returned.
Otherwise, the [filter](https://cloud.google.com/monitoring/api/v3/filters) specifies which metric descriptors are to be returned. For example, the following filter matches all [custom metrics](https://cloud.google.com/monitoring/custom-metrics):

metric.type = starts_with("custom.googleapis.com/")
`),
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

	// Add list_available_metrics tool
	listAvailableMetricsTool := mcp.NewTool("list_available_metrics",
		mcp.WithDescription("List available metrics in Cloud Monitoring"),
		mcp.WithString("filter",
			mcp.Description(`Filter expression for metric descriptors.
If this field is empty, all custom and system-defined metric descriptors are returned.
Otherwise, the [filter](https://cloud.google.com/monitoring/api/v3/filters) specifies which metric descriptors are to be returned. For example, the following filter matches all [custom metrics](https://cloud.google.com/monitoring/custom-metrics):

metric.type = starts_with("custom.googleapis.com/")
`),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Maximum number of metrics to return (default: 100)"),
		),
		mcp.WithString("page_token",
			mcp.Description("Page token for pagination"),
		),
	)

	// Add list_traces tool
	listTracesTool := mcp.NewTool("list_traces",
		mcp.WithDescription("List traces from Cloud Trace"),
		mcp.WithString("start_time",
			mcp.Required(),
			mcp.Description("Start time for the query (ISO 8601 format)"),
		),
		mcp.WithString("end_time",
			mcp.Required(),
			mcp.Description("End time for the query (ISO 8601 format)"),
		),
		mcp.WithString("filter",
			mcp.Description(`By default, searches use prefix matching. To specify exact match, prepend
  a plus symbol (+) to the search term.
  Multiple terms are ANDed. Syntax:

    - root:NAME_PREFIX or NAME_PREFIX: Return traces where any root
      span starts with NAME_PREFIX.
    - +root:NAME or +NAME: Return traces where any root span's name is
      exactly NAME.
    - span:NAME_PREFIX: Return traces where any span starts with
      NAME_PREFIX.
    - +span:NAME: Return traces where any span's name is exactly
      NAME.
    - latency:DURATION: Return traces whose overall latency is
      greater or equal to than DURATION. Accepted units are nanoseconds
      (ns), milliseconds (ms), and seconds (s). Default is ms. For
      example, latency:24ms returns traces whose overall latency
      is greater than or equal to 24 milliseconds.
    - label:LABEL_KEY: Return all traces containing the specified
      label key (exact match, case-sensitive) regardless of the key:value
      pair's value (including empty values).
    - LABEL_KEY:VALUE_PREFIX: Return all traces containing the specified
      label key (exact match, case-sensitive) whose value starts with
      VALUE_PREFIX. Both a key and a value must be specified.
    - +LABEL_KEY:VALUE: Return all traces containing a key:value pair
      exactly matching the specified text. Both a key and a value must be
      specified.
    - method:VALUE: Equivalent to /http/method:VALUE.
    - url:VALUE: Equivalent to /http/url:VALUE.
      `),
		),
		mcp.WithString("order_by",
			mcp.Description("Order by field (e.g., 'start_time desc')"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Maximum number of traces to return (default: 100)"),
		),
		mcp.WithString("page_token",
			mcp.Description("Page token for pagination"),
		),
	)

	// Add get_trace tool
	getTraceTool := mcp.NewTool("get_trace",
		mcp.WithDescription("Get a specific trace from Cloud Trace"),
		mcp.WithString("trace_id",
			mcp.Required(),
			mcp.Description("Trace ID to retrieve"),
		),
	)

	// Add patch_traces tool
	patchTracesTool := mcp.NewTool("patch_traces",
		mcp.WithDescription("Update trace spans in Cloud Trace"),
		mcp.WithString("trace_id",
			mcp.Required(),
			mcp.Description("Trace ID to update"),
		),
		mcp.WithObject("spans",
			mcp.Required(),
			mcp.Description("Array of span objects to update or create"),
		),
	)

	// Add create_profile tool
	createProfileTool := mcp.NewTool("create_profile",
		mcp.WithDescription("Create a new profile in Cloud Profiler"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target deployment name"),
		),
		mcp.WithString("profile_type",
			mcp.Required(),
			mcp.Description("Profile type: CPU, HEAP, THREADS, CONTENTION, or WALL"),
		),
		mcp.WithString("duration",
			mcp.Description("Profile duration (e.g., '60s', '5m', defaults to '60s')"),
		),
		mcp.WithObject("labels",
			mcp.Description("Optional labels for the profile"),
		),
	)

	// Add create_offline_profile tool
	createOfflineProfileTool := mcp.NewTool("create_offline_profile",
		mcp.WithDescription("Create an offline profile in Cloud Profiler"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target deployment name"),
		),
		mcp.WithString("profile_type",
			mcp.Required(),
			mcp.Description("Profile type: CPU, HEAP, THREADS, CONTENTION, or WALL"),
		),
		mcp.WithString("profile_data",
			mcp.Required(),
			mcp.Description("Base64-encoded profile data"),
		),
		mcp.WithString("duration",
			mcp.Description("Profile duration (e.g., '60s', '5m')"),
		),
		mcp.WithObject("labels",
			mcp.Description("Optional labels for the profile"),
		),
	)

	// Add update_profile tool
	updateProfileTool := mcp.NewTool("update_profile",
		mcp.WithDescription("Update a profile in Cloud Profiler"),
		mcp.WithString("profile_name",
			mcp.Required(),
			mcp.Description("Profile name to update"),
		),
		mcp.WithString("profile_data",
			mcp.Description("Updated base64-encoded profile data"),
		),
		mcp.WithObject("labels",
			mcp.Description("Updated labels for the profile"),
		),
		mcp.WithString("update_mask",
			mcp.Description("Fields to update (e.g., 'labels,profile_bytes')"),
		),
	)

	// Add list_profiles tool
	listProfilesTool := mcp.NewTool("list_profiles",
		mcp.WithDescription("List profiles from Cloud Profiler"),
		mcp.WithNumber("page_size",
			mcp.Description("Maximum number of profiles to return (default: 100)"),
		),
		mcp.WithString("page_token",
			mcp.Description("Page token for pagination"),
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
	s.AddTool(listAvailableMetricsTool, createListAvailableMetricsHandler(monitoringClient))
	s.AddTool(listTracesTool, createListTracesHandler(traceClient))
	s.AddTool(getTraceTool, createGetTraceHandler(traceClient))
	s.AddTool(patchTracesTool, createPatchTracesHandler(traceClient))
	s.AddTool(createProfileTool, createProfileHandler(profilerClient))
	s.AddTool(createOfflineProfileTool, createOfflineProfileHandler(profilerClient))
	s.AddTool(updateProfileTool, updateProfileHandler(profilerClient))
	s.AddTool(listProfilesTool, listProfilesHandler(profilerClient))

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

// createListAvailableMetricsHandler creates a handler for listing available metrics
func createListAvailableMetricsHandler(client monitoring.MonitoringClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		req := monitoring.ListAvailableMetricsRequest{
			PageSize: 100, // default
		}

		// Parse optional filter parameter
		if filterArg, exists := args["filter"]; exists {
			if filter, ok := filterArg.(string); ok && filter != "" {
				req.Filter = filter
			}
		}

		// Parse optional page_size parameter
		if pageSizeArg, exists := args["page_size"]; exists {
			if pageSize, ok := pageSizeArg.(float64); ok && pageSize > 0 {
				req.PageSize = int(pageSize)
			}
		}

		// Parse optional page_token parameter
		if pageTokenArg, exists := args["page_token"]; exists {
			if pageToken, ok := pageTokenArg.(string); ok && pageToken != "" {
				req.PageToken = pageToken
			}
		}

		metrics, err := client.ListAvailableMetrics(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list available metrics: %v", err)), nil
		}

		// Convert metrics to JSON for response
		metricsJSON, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal available metrics: %v", err)), nil
		}

		return mcp.NewToolResultText(string(metricsJSON)), nil
	}
}

// createListTracesHandler creates a handler for listing traces
func createListTracesHandler(client trace.TraceClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		args := request.GetArguments()
		req := trace.ListTracesRequest{
			StartTime: startTime,
			EndTime:   endTime,
			PageSize:  100, // default
		}

		// Parse optional filter parameter
		if filterArg, exists := args["filter"]; exists {
			if filter, ok := filterArg.(string); ok && filter != "" {
				req.Filter = filter
			}
		}

		// Parse optional order_by parameter
		if orderByArg, exists := args["order_by"]; exists {
			if orderBy, ok := orderByArg.(string); ok && orderBy != "" {
				req.OrderBy = orderBy
			}
		}

		// Parse optional page_size parameter
		if pageSizeArg, exists := args["page_size"]; exists {
			if pageSize, ok := pageSizeArg.(float64); ok && pageSize > 0 {
				req.PageSize = int(pageSize)
			}
		}

		// Parse optional page_token parameter
		if pageTokenArg, exists := args["page_token"]; exists {
			if pageToken, ok := pageTokenArg.(string); ok && pageToken != "" {
				req.PageToken = pageToken
			}
		}

		traces, err := client.ListTraces(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list traces: %v", err)), nil
		}

		// Convert traces to JSON for response
		tracesJSON, err := json.MarshalIndent(traces, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal traces: %v", err)), nil
		}

		return mcp.NewToolResultText(string(tracesJSON)), nil
	}
}

// createGetTraceHandler creates a handler for getting a specific trace
func createGetTraceHandler(client trace.TraceClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		traceID, err := request.RequireString("trace_id")
		if err != nil {
			return mcp.NewToolResultError("trace_id is required"), nil
		}

		req := trace.GetTraceRequest{
			TraceID: traceID,
		}

		traceResult, err := client.GetTrace(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get trace: %v", err)), nil
		}

		// Convert trace to JSON for response
		traceJSON, err := json.MarshalIndent(traceResult, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal trace: %v", err)), nil
		}

		return mcp.NewToolResultText(string(traceJSON)), nil
	}
}

// createPatchTracesHandler creates a handler for updating trace spans
func createPatchTracesHandler(client trace.TraceClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		traceID, err := request.RequireString("trace_id")
		if err != nil {
			return mcp.NewToolResultError("trace_id is required"), nil
		}

		args := request.GetArguments()
		spansArg, exists := args["spans"]
		if !exists {
			return mcp.NewToolResultError("spans is required"), nil
		}

		// Parse spans from the request
		var spans []trace.Span
		if spansArray, ok := spansArg.([]any); ok {
			for _, spanData := range spansArray {
				if spanObj, ok := spanData.(map[string]any); ok {
					span := trace.Span{}

					if spanID, ok := spanObj["span_id"].(string); ok {
						span.SpanID = spanID
					}

					if name, ok := spanObj["name"].(string); ok {
						span.Name = name
					}

					if parentID, ok := spanObj["parent_id"].(string); ok {
						span.ParentID = parentID
					}

					if kind, ok := spanObj["kind"].(string); ok {
						span.Kind = kind
					}

					// Parse start_time
					if startTimeStr, ok := spanObj["start_time"].(string); ok {
						if startTime, parseErr := time.Parse(time.RFC3339, startTimeStr); parseErr == nil {
							span.StartTime = startTime
						}
					}

					// Parse end_time
					if endTimeStr, ok := spanObj["end_time"].(string); ok {
						if endTime, parseErr := time.Parse(time.RFC3339, endTimeStr); parseErr == nil {
							span.EndTime = endTime
						}
					}

					// Parse labels
					if labelsObj, ok := spanObj["labels"].(map[string]any); ok {
						span.Labels = make(map[string]string)
						for k, v := range labelsObj {
							if str, ok := v.(string); ok {
								span.Labels[k] = str
							}
						}
					}

					spans = append(spans, span)
				}
			}
		} else {
			return mcp.NewToolResultError("spans must be an array of span objects"), nil
		}

		req := trace.PatchTraceRequest{
			TraceID: traceID,
			Spans:   spans,
		}

		err = client.PatchTraces(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to patch traces: %v", err)), nil
		}

		return mcp.NewToolResultText("Trace spans updated successfully"), nil
	}
}

// createProfileHandler creates a handler for creating profiles
func createProfileHandler(client profiler.ProfilerClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		target, err := request.RequireString("target")
		if err != nil {
			return mcp.NewToolResultError("target is required"), nil
		}

		profileTypeStr, err := request.RequireString("profile_type")
		if err != nil {
			return mcp.NewToolResultError("profile_type is required"), nil
		}

		args := request.GetArguments()
		duration := "60s" // default
		if durationArg, exists := args["duration"]; exists {
			if d, ok := durationArg.(string); ok && d != "" {
				duration = d
			}
		}

		// Parse labels
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsObj, ok := labelsArg.(map[string]any); ok {
				labels = make(map[string]string)
				for k, v := range labelsObj {
					if str, ok := v.(string); ok {
						labels[k] = str
					}
				}
			}
		}

		req := profiler.CreateProfileRequest{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			Deployment: &profiler.Deployment{
				ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
				Target:    target,
				Labels:    labels,
			},
			ProfileType: []profiler.ProfileType{profiler.ProfileType(profileTypeStr)},
			Duration:    duration,
			Labels:      labels,
		}

		profile, err := client.CreateProfile(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create profile: %v", err)), nil
		}

		// Convert profile to JSON for response
		profileJSON, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal profile: %v", err)), nil
		}

		return mcp.NewToolResultText(string(profileJSON)), nil
	}
}

// createOfflineProfileHandler creates a handler for creating offline profiles
func createOfflineProfileHandler(client profiler.ProfilerClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		target, err := request.RequireString("target")
		if err != nil {
			return mcp.NewToolResultError("target is required"), nil
		}

		profileTypeStr, err := request.RequireString("profile_type")
		if err != nil {
			return mcp.NewToolResultError("profile_type is required"), nil
		}

		profileData, err := request.RequireString("profile_data")
		if err != nil {
			return mcp.NewToolResultError("profile_data is required"), nil
		}

		args := request.GetArguments()
		duration := "60s" // default
		if durationArg, exists := args["duration"]; exists {
			if d, ok := durationArg.(string); ok && d != "" {
				duration = d
			}
		}

		// Parse labels
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsObj, ok := labelsArg.(map[string]any); ok {
				labels = make(map[string]string)
				for k, v := range labelsObj {
					if str, ok := v.(string); ok {
						labels[k] = str
					}
				}
			}
		}

		req := profiler.CreateOfflineProfileRequest{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			Profile: &profiler.Profile{
				ProfileType:  profiler.ProfileType(profileTypeStr),
				Duration:     duration,
				Labels:       labels,
				ProfileBytes: profileData,
				Deployment: &profiler.Deployment{
					ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
					Target:    target,
					Labels:    labels,
				},
			},
		}

		profile, err := client.CreateOfflineProfile(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create offline profile: %v", err)), nil
		}

		// Convert profile to JSON for response
		profileJSON, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal profile: %v", err)), nil
		}

		return mcp.NewToolResultText(string(profileJSON)), nil
	}
}

// updateProfileHandler creates a handler for updating profiles
func updateProfileHandler(client profiler.ProfilerClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		profileName, err := request.RequireString("profile_name")
		if err != nil {
			return mcp.NewToolResultError("profile_name is required"), nil
		}

		args := request.GetArguments()
		var profileData string
		if profileDataArg, exists := args["profile_data"]; exists {
			if pd, ok := profileDataArg.(string); ok {
				profileData = pd
			}
		}

		var updateMask string
		if updateMaskArg, exists := args["update_mask"]; exists {
			if um, ok := updateMaskArg.(string); ok {
				updateMask = um
			}
		}

		// Parse labels
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsObj, ok := labelsArg.(map[string]any); ok {
				labels = make(map[string]string)
				for k, v := range labelsObj {
					if str, ok := v.(string); ok {
						labels[k] = str
					}
				}
			}
		}

		req := profiler.UpdateProfileRequest{
			Profile: &profiler.Profile{
				Name:   profileName,
				Labels: labels,
			},
			ProfileBytes: profileData,
			UpdateMask:   updateMask,
		}

		profile, err := client.UpdateProfile(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update profile: %v", err)), nil
		}

		// Convert profile to JSON for response
		profileJSON, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal profile: %v", err)), nil
		}

		return mcp.NewToolResultText(string(profileJSON)), nil
	}
}

// listProfilesHandler creates a handler for listing profiles
func listProfilesHandler(client profiler.ProfilerClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		req := profiler.ListProfilesRequest{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			PageSize:  100, // default
		}

		// Parse optional page_size parameter
		if pageSizeArg, exists := args["page_size"]; exists {
			if pageSize, ok := pageSizeArg.(float64); ok && pageSize > 0 {
				req.PageSize = int64(pageSize)
			}
		}

		// Parse optional page_token parameter
		if pageTokenArg, exists := args["page_token"]; exists {
			if pageToken, ok := pageTokenArg.(string); ok && pageToken != "" {
				req.PageToken = pageToken
			}
		}

		profiles, err := client.ListProfiles(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list profiles: %v", err)), nil
		}

		// Convert profiles to JSON for response
		profilesJSON, err := json.MarshalIndent(profiles, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal profiles: %v", err)), nil
		}

		return mcp.NewToolResultText(string(profilesJSON)), nil
	}
}
