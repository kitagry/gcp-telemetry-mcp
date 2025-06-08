package logging

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/logging"
)

func TestCloudLoggingClient_WriteEntry(t *testing.T) {
	tests := []struct {
		name    string
		entry   LogEntry
		wantErr bool
	}{
		{
			name: "write simple log entry",
			entry: LogEntry{
				Severity: "INFO",
				Message:  "test message",
				Labels: map[string]string{
					"source": "test",
				},
			},
			wantErr: false,
		},
		{
			name: "write log entry with payload",
			entry: LogEntry{
				Severity: "ERROR",
				Message:  "error occurred",
				Payload: map[string]any{
					"error_code": 500,
					"details":    "internal server error",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockCloudLoggingClient{}
			client := &CloudLoggingClient{
				client: mockClient,
			}

			err := client.WriteEntry(context.Background(), "test-log", tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(mockClient.entries) != 1 {
				t.Errorf("Expected 1 entry to be written, got %d", len(mockClient.entries))
			}
		})
	}
}

func TestCloudLoggingClient_ListEntries(t *testing.T) {
	mockClient := &MockCloudLoggingClient{
		entries: []logging.Entry{
			{
				Timestamp: time.Now(),
				Severity:  logging.Info,
				Payload:   "test message 1",
			},
			{
				Timestamp: time.Now(),
				Severity:  logging.Error,
				Payload:   "test message 2",
			},
		},
	}

	client := &CloudLoggingClient{
		client: mockClient,
	}

	entries, err := client.ListEntries(context.Background(), ListEntriesRequest{
		Filter: "severity>=INFO",
		Limit:  10,
	})
	if err != nil {
		t.Errorf("ListEntries() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

// MockCloudLoggingClient implements the logging client interface for testing
type MockCloudLoggingClient struct {
	entries []logging.Entry
}

func (m *MockCloudLoggingClient) WriteEntry(ctx context.Context, logName string, entry LogEntry) error {
	// Convert LogEntry to logging.Entry for storage
	logEntry := logging.Entry{
		Timestamp: time.Now(),
		Payload:   entry.Message,
		Labels:    entry.Labels,
	}

	// Convert severity string to logging.Severity
	switch entry.Severity {
	case "DEBUG":
		logEntry.Severity = logging.Debug
	case "INFO":
		logEntry.Severity = logging.Info
	case "WARNING":
		logEntry.Severity = logging.Warning
	case "ERROR":
		logEntry.Severity = logging.Error
	default:
		logEntry.Severity = logging.Info
	}

	if entry.Payload != nil {
		logEntry.Payload = entry.Payload
	}

	m.entries = append(m.entries, logEntry)
	return nil
}

func (m *MockCloudLoggingClient) ListEntries(ctx context.Context, req ListEntriesRequest) ([]LogEntry, error) {
	var result []LogEntry
	for _, entry := range m.entries {
		logEntry := LogEntry{
			Labels:    entry.Labels,
			Timestamp: entry.Timestamp,
		}

		// Handle payload safely
		if strPayload, ok := entry.Payload.(string); ok {
			logEntry.Message = strPayload
		} else if payload, ok := entry.Payload.(map[string]any); ok {
			logEntry.Payload = payload
		}

		// Convert severity to string
		switch entry.Severity {
		case logging.Debug:
			logEntry.Severity = "DEBUG"
		case logging.Info:
			logEntry.Severity = "INFO"
		case logging.Warning:
			logEntry.Severity = "WARNING"
		case logging.Error:
			logEntry.Severity = "ERROR"
		default:
			logEntry.Severity = "INFO"
		}

		result = append(result, logEntry)
	}
	return result, nil
}
