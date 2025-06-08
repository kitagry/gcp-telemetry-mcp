package logging

//go:generate go tool mockgen -destination=mocks/mock_client.go -package=mocks github.com/kitagry/gcp-telemetry-mcp/logging LoggingClient,LoggingClientInterface

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
)

// LogEntry represents a log entry to be written or retrieved
type LogEntry struct {
	Severity  string            `json:"severity"`
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels,omitempty"`
	Payload   map[string]any    `json:"payload,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// ListEntriesRequest represents a request to list log entries
type ListEntriesRequest struct {
	Filter    string `json:"filter,omitempty"`
	OrderBy   string `json:"order_by,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

// LoggingClient defines the interface for Cloud Logging operations
type LoggingClient interface {
	WriteEntry(ctx context.Context, logName string, entry LogEntry) error
	ListEntries(ctx context.Context, req ListEntriesRequest) ([]LogEntry, error)
}

// CloudLoggingClient implements LoggingClient using Google Cloud Logging
type CloudLoggingClient struct {
	client LoggingClientInterface
}

// LoggingClientInterface abstracts the Google Cloud Logging client for testing
type LoggingClientInterface interface {
	WriteEntry(ctx context.Context, logName string, entry LogEntry) error
	ListEntries(ctx context.Context, req ListEntriesRequest) ([]LogEntry, error)
}

// New creates a new CloudLoggingClient
func New(projectID string) (*CloudLoggingClient, error) {
	client, err := logging.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, err
	}

	adminClient, err := logadmin.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, err
	}

	return &CloudLoggingClient{
		client: &realLoggingClient{
			client:      client,
			adminClient: adminClient,
		},
	}, nil
}

// NewWithClient creates a new CloudLoggingClient with a custom interface for testing
func NewWithClient(client LoggingClientInterface) *CloudLoggingClient {
	return &CloudLoggingClient{
		client: client,
	}
}

// WriteEntry writes a log entry to Cloud Logging
func (c *CloudLoggingClient) WriteEntry(ctx context.Context, logName string, entry LogEntry) error {
	return c.client.WriteEntry(ctx, logName, entry)
}

// ListEntries retrieves log entries from Cloud Logging
func (c *CloudLoggingClient) ListEntries(ctx context.Context, req ListEntriesRequest) ([]LogEntry, error) {
	return c.client.ListEntries(ctx, req)
}

// realLoggingClient wraps the actual Google Cloud Logging client
type realLoggingClient struct {
	client      *logging.Client
	adminClient *logadmin.Client
}

// WriteEntry implements LoggingClientInterface for the real client
func (r *realLoggingClient) WriteEntry(ctx context.Context, logName string, entry LogEntry) error {
	logger := r.client.Logger(logName)
	defer logger.Flush()

	// Convert severity string to logging.Severity
	var severity logging.Severity
	switch entry.Severity {
	case "DEBUG":
		severity = logging.Debug
	case "INFO":
		severity = logging.Info
	case "WARNING":
		severity = logging.Warning
	case "ERROR":
		severity = logging.Error
	case "CRITICAL":
		severity = logging.Critical
	default:
		severity = logging.Info
	}

	logEntry := logging.Entry{
		Severity: severity,
		Labels:   entry.Labels,
	}

	// Set payload - prefer structured payload over message
	if entry.Payload != nil {
		logEntry.Payload = entry.Payload
	} else {
		logEntry.Payload = entry.Message
	}

	logger.Log(logEntry)
	return nil
}

// ListEntries implements LoggingClientInterface for the real client
func (r *realLoggingClient) ListEntries(ctx context.Context, req ListEntriesRequest) ([]LogEntry, error) {
	// Set limit, default to 50 if not specified
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	// Create an iterator for log entries using the admin client
	iterator := r.adminClient.Entries(ctx, logadmin.Filter(req.Filter), logadmin.NewestFirst())

	var entries []LogEntry
	count := 0

	// Iterate through the entries
	for count < limit {
		entry, err := iterator.Next()
		if err != nil {
			// Check for iterator done using Google API standard approach
			if err.Error() == "no more items in iterator" {
				break
			}
			return nil, err
		}

		// Convert logging.Entry to our LogEntry format
		logEntry := LogEntry{
			Timestamp: entry.Timestamp,
			Labels:    entry.Labels,
		}

		// Convert severity
		switch entry.Severity {
		case logging.Debug:
			logEntry.Severity = "DEBUG"
		case logging.Info:
			logEntry.Severity = "INFO"
		case logging.Warning:
			logEntry.Severity = "WARNING"
		case logging.Error:
			logEntry.Severity = "ERROR"
		case logging.Critical:
			logEntry.Severity = "CRITICAL"
		default:
			logEntry.Severity = "INFO"
		}

		// Handle payload - could be string or structured data
		if entry.Payload != nil {
			switch payload := entry.Payload.(type) {
			case string:
				logEntry.Message = payload
			case map[string]any:
				logEntry.Payload = payload
				// Try to extract message from payload if available
				if msg, ok := payload["message"]; ok {
					if msgStr, ok := msg.(string); ok {
						logEntry.Message = msgStr
					}
				}
			default:
				// Convert other types to string
				logEntry.Message = fmt.Sprintf("%v", payload)
			}
		}

		entries = append(entries, logEntry)
		count++
	}

	return entries, nil
}
