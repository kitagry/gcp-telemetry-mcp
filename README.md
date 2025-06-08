# GCP Telemetry MCP Server

A Model Context Protocol (MCP) server for Google Cloud Platform telemetry services, providing seamless integration with GCP observability tools.

## Features

### Cloud Logging
- ✅ Write log entries with structured data
- ✅ Support for multiple severity levels (DEBUG, INFO, WARNING, ERROR, CRITICAL)
- ✅ Custom labels and structured payloads
- ✅ List log entries with filtering and pagination

### Cloud Monitoring
- ✅ Create custom metric descriptors
- ✅ Write time series data points
- ✅ Query time series data with advanced filtering
- ✅ Support for all metric kinds (GAUGE, DELTA, CUMULATIVE)
- ✅ Support for all value types (BOOL, INT64, DOUBLE, STRING, DISTRIBUTION)
- ✅ Advanced aggregation options (alignment periods, reducers)
- ✅ Delete custom metric descriptors
- ✅ List available metric descriptors

### Planned Features
- 🔄 Cloud Trace
- 🔄 Cloud Profiler

## Prerequisites

- Go 1.24.2 or later
- Google Cloud Project with appropriate APIs enabled
- Authentication credentials configured (see [Authentication](#authentication))

## Installation

### From Source

```bash
git clone https://github.com/kitagry/gcp-telemetry-mcp.git
cd gcp-telemetry-mcp
go build -o gcp-telemetry-mcp
```

## Authentication

This server uses Google Cloud authentication. Configure one of the following:

1. **Service Account Key File**:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
   ```

2. **Application Default Credentials** (if running on GCP):
   ```bash
   gcloud auth application-default login
   ```

3. **Workload Identity** (for GKE/Cloud Run deployments)

## Configuration

Set the required environment variable:

```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

## Usage

### Running the Server

```bash
./gcp-telemetry-mcp
```

Or with Go:

```bash
go run main.go
```

### MCP Tools

The server provides the following MCP tools:

## Cloud Logging Tools

#### `write_log_entry`

Write a log entry to Cloud Logging.

**Parameters:**
- `log_name` (string, required): Name of the log to write to
- `severity` (string, required): Log severity (DEBUG, INFO, WARNING, ERROR, CRITICAL)
- `message` (string, required): Log message
- `labels` (object, optional): Key-value pairs for log labels
- `payload` (object, optional): Structured data payload

**Example:**
```json
{
  "log_name": "my-application-log",
  "severity": "INFO",
  "message": "User login successful",
  "labels": {
    "user_id": "12345",
    "environment": "production"
  },
  "payload": {
    "event": "user_login",
    "timestamp": "2024-01-01T12:00:00Z",
    "ip_address": "192.168.1.1"
  }
}
```

#### `list_log_entries`

List log entries from Cloud Logging.

**Parameters:**
- `filter` (string, optional): Cloud Logging filter expression
- `limit` (number, optional): Maximum number of entries to return (default: 50)

**Example:**
```json
{
  "filter": "severity>=ERROR",
  "limit": 100
}
```

## Cloud Monitoring Tools

#### `create_metric_descriptor`

Create a custom metric descriptor in Cloud Monitoring.

**Parameters:**
- `type` (string, required): Metric type (e.g., 'custom.googleapis.com/my_metric')
- `metric_kind` (string, required): Metric kind (GAUGE, DELTA, or CUMULATIVE)
- `value_type` (string, required): Value type (BOOL, INT64, DOUBLE, STRING, or DISTRIBUTION)
- `description` (string, required): Description of the metric
- `display_name` (string, optional): Display name for the metric

**Example:**
```json
{
  "type": "custom.googleapis.com/app/response_time",
  "metric_kind": "GAUGE",
  "value_type": "DOUBLE",
  "description": "Application response time in seconds",
  "display_name": "App Response Time"
}
```

#### `write_time_series`

Write time series data to Cloud Monitoring.

**Parameters:**
- `metric_type` (string, required): Metric type to write data for
- `resource_type` (string, required): Resource type (e.g., 'global', 'gce_instance')
- `value` (number, required): Metric value to write
- `metric_labels` (object, optional): Optional metric labels
- `timestamp` (string, optional): Timestamp for the data point (ISO 8601 format, defaults to now)

**Example:**
```json
{
  "metric_type": "custom.googleapis.com/app/response_time",
  "resource_type": "global",
  "value": 0.125,
  "metric_labels": {
    "service": "api",
    "version": "v1.2.0"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### `list_time_series`

List time series data from Cloud Monitoring.

**Parameters:**
- `filter` (string, required): Monitoring filter expression
- `start_time` (string, required): Start time for the query (ISO 8601 format)
- `end_time` (string, required): End time for the query (ISO 8601 format)
- `aggregation` (object, optional): Aggregation configuration

**Example:**
```json
{
  "filter": "metric.type=\"compute.googleapis.com/instance/cpu/usage\"",
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": "2024-01-01T12:00:00Z",
  "aggregation": {
    "alignment_period": "60s",
    "per_series_aligner": "ALIGN_MEAN",
    "cross_series_reducer": "REDUCE_MEAN",
    "group_by_fields": ["resource.zone"]
  }
}
```

#### `list_metric_descriptors`

List metric descriptors from Cloud Monitoring.

**Parameters:**
- `filter` (string, optional): Filter expression for metric descriptors

**Example:**
```json
{
  "filter": "metric.type=starts_with(\"custom.googleapis.com/\")"
}
```

#### `delete_metric_descriptor`

Delete a custom metric descriptor from Cloud Monitoring.

**Parameters:**
- `metric_type` (string, required): Metric type to delete

**Example:**
```json
{
  "metric_type": "custom.googleapis.com/my_old_metric"
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── main.go              # MCP server implementation and tool handlers
├── logging/
│   ├── client.go        # Cloud Logging client implementation
│   └── client_test.go   # Tests for logging client
├── monitoring/
│   ├── client.go        # Cloud Monitoring client implementation
│   └── client_test.go   # Tests for monitoring client
├── go.mod               # Go module definition
├── go.sum               # Go dependency checksums
└── README.md           # This file
```

## Error Handling

The server provides detailed error messages for common issues:

- Missing `GOOGLE_CLOUD_PROJECT` environment variable
- Authentication failures
- Invalid parameters for logging and monitoring operations
- Cloud Logging API errors
- Cloud Monitoring API errors
- Time series data validation errors
- Metric descriptor creation/deletion failures

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- Create an issue in the GitHub repository
- Check Google Cloud Logging documentation for logging-specific questions
- Check Google Cloud Monitoring documentation for monitoring-specific questions
