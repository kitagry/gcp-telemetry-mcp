# GCP Telemetry MCP Server

A Model Context Protocol (MCP) server for Google Cloud Platform telemetry services, providing seamless integration with GCP observability tools.

## Features

### Cloud Logging
- ✅ Write log entries with structured data
- ✅ Support for multiple severity levels (DEBUG, INFO, WARNING, ERROR, CRITICAL)
- ✅ Custom labels and structured payloads
- ✅ List log entries (basic implementation)

### Planned Features
- 🔄 Cloud Monitoring
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
./cloud-logging-mcp
```

Or with Go:

```bash
go run main.go
```

### MCP Tools

The server provides the following MCP tools:

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

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── main.go              # MCP server implementation
├── logging/
│   ├── client.go        # Cloud Logging client implementation
│   └── client_test.go   # Tests for logging client
├── go.mod               # Go module definition
└── README.md           # This file
```

## Error Handling

The server provides detailed error messages for common issues:

- Missing `GOOGLE_CLOUD_PROJECT` environment variable
- Authentication failures
- Invalid log parameters
- Cloud Logging API errors

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
- Check Google Cloud Logging documentation for API-specific questions
