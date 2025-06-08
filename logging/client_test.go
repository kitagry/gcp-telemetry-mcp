package logging_test

import (
	"context"
	"testing"
	"time"

	"github.com/kitagry/gcp-telemetry-mcp/logging"
	"github.com/kitagry/gcp-telemetry-mcp/logging/mocks"
	"go.uber.org/mock/gomock"
)

func TestCloudLoggingClient_WriteEntry(t *testing.T) {
	tests := []struct {
		name    string
		entry   logging.LogEntry
		wantErr bool
	}{
		{
			name: "write simple log entry",
			entry: logging.LogEntry{
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
			entry: logging.LogEntry{
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockLoggingClientInterface(ctrl)
			client := logging.NewWithClient(mockClient)

			// Set expectation for WriteEntry call
			mockClient.EXPECT().
				WriteEntry(gomock.Any(), "test-log", tt.entry).
				Return(nil).
				Times(1)

			err := client.WriteEntry(context.Background(), "test-log", tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloudLoggingClient_ListEntries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedEntries := []logging.LogEntry{
		{
			Timestamp: time.Now(),
			Severity:  "INFO",
			Message:   "test message 1",
		},
		{
			Timestamp: time.Now(),
			Severity:  "ERROR",
			Message:   "test message 2",
		},
	}

	mockClient := mocks.NewMockLoggingClientInterface(ctrl)
	client := logging.NewWithClient(mockClient)

	req := logging.ListEntriesRequest{
		Filter: "severity>=INFO",
		Limit:  10,
	}

	// Set expectation for ListEntries call
	mockClient.EXPECT().
		ListEntries(gomock.Any(), req).
		Return(expectedEntries, nil).
		Times(1)

	entries, err := client.ListEntries(context.Background(), req)
	if err != nil {
		t.Errorf("ListEntries() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Severity != "INFO" {
		t.Errorf("Expected first entry severity to be INFO, got %s", entries[0].Severity)
	}

	if entries[1].Severity != "ERROR" {
		t.Errorf("Expected second entry severity to be ERROR, got %s", entries[1].Severity)
	}
}
