package trace_test

import (
	"context"
	"testing"
	"time"

	"github.com/kitagry/gcp-telemetry-mcp/trace"
	"github.com/kitagry/gcp-telemetry-mcp/trace/mocks"
	"go.uber.org/mock/gomock"
)

func TestCloudTraceClient_ListTraces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedTraces := []trace.Trace{
		{
			TraceID:   "trace123",
			ProjectID: "test-project",
			Spans: []trace.Span{
				{
					SpanID:    "span123",
					Name:      "test-span",
					StartTime: time.Now().Add(-1 * time.Hour),
					EndTime:   time.Now(),
					Kind:      "RPC_SERVER",
					Labels: map[string]string{
						"component": "test",
					},
				},
			},
		},
	}

	mockClient := mocks.NewMockTraceClientInterface(ctrl)
	client := trace.NewWithClient(mockClient, "test-project")

	req := trace.ListTracesRequest{
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now(),
		Filter:    "span_name_prefix:\"test\"",
		PageSize:  50,
	}

	// Set expectation for ListTraces call
	mockClient.EXPECT().
		ListTraces(gomock.Any(), req).
		Return(expectedTraces, nil).
		Times(1)

	result, err := client.ListTraces(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(result))
	}

	if result[0].TraceID != expectedTraces[0].TraceID {
		t.Errorf("Expected trace ID %s, got %s", expectedTraces[0].TraceID, result[0].TraceID)
	}

	if len(result[0].Spans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(result[0].Spans))
	}
}

func TestCloudTraceClient_GetTrace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedTrace := &trace.Trace{
		TraceID:   "trace123",
		ProjectID: "test-project",
		Spans: []trace.Span{
			{
				SpanID:    "span123",
				Name:      "test-span",
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
				Kind:      "RPC_SERVER",
			},
		},
	}

	mockClient := mocks.NewMockTraceClientInterface(ctrl)
	client := trace.NewWithClient(mockClient, "test-project")

	req := trace.GetTraceRequest{
		TraceID: "trace123",
	}

	// Set expectation for GetTrace call
	mockClient.EXPECT().
		GetTrace(gomock.Any(), req).
		Return(expectedTrace, nil).
		Times(1)

	result, err := client.GetTrace(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.TraceID != expectedTrace.TraceID {
		t.Errorf("Expected trace ID %s, got %s", expectedTrace.TraceID, result.TraceID)
	}

	if len(result.Spans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(result.Spans))
	}
}

func TestCloudTraceClient_PatchTraces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTraceClientInterface(ctrl)
	client := trace.NewWithClient(mockClient, "test-project")

	req := trace.PatchTraceRequest{
		TraceID: "trace123",
		Spans: []trace.Span{
			{
				SpanID:    "span123",
				Name:      "updated-span",
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
				Kind:      "RPC_CLIENT",
				Labels: map[string]string{
					"updated": "true",
				},
			},
		},
	}

	// Set expectation for PatchTraces call
	mockClient.EXPECT().
		PatchTraces(gomock.Any(), req).
		Return(nil).
		Times(1)

	err := client.PatchTraces(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}