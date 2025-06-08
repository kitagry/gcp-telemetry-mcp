package monitoring_test

import (
	"context"
	"testing"
	"time"

	"github.com/kitagry/gcp-telemetry-mcp/monitoring"
	"github.com/kitagry/gcp-telemetry-mcp/monitoring/mocks"
	"go.uber.org/mock/gomock"
)

func TestCloudMonitoringClient_CreateMetricDescriptor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMonitoringClientInterface(ctrl)
	client := monitoring.NewWithClient(mockClient, "test-project")

	req := monitoring.CreateMetricRequest{
		MetricDescriptor: monitoring.MetricDescriptor{
			Type:        "custom.googleapis.com/test_metric",
			MetricKind:  "GAUGE",
			ValueType:   "DOUBLE",
			Description: "Test metric",
			DisplayName: "Test Metric",
		},
	}

	// Set expectation for CreateMetricDescriptor call
	mockClient.EXPECT().
		CreateMetricDescriptor(gomock.Any(), req).
		Return(nil).
		Times(1)

	err := client.CreateMetricDescriptor(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCloudMonitoringClient_WriteTimeSeries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMonitoringClientInterface(ctrl)
	client := monitoring.NewWithClient(mockClient, "test-project")

	req := monitoring.WriteTimeSeriesRequest{
		TimeSeries: []monitoring.TimeSeriesData{
			{
				MetricType:   "custom.googleapis.com/test_metric",
				ResourceType: "global",
				Values: []monitoring.MetricValue{
					{
						Value:     42.0,
						Timestamp: time.Now(),
					},
				},
			},
		},
	}

	// Set expectation for WriteTimeSeries call
	mockClient.EXPECT().
		WriteTimeSeries(gomock.Any(), req).
		Return(nil).
		Times(1)

	err := client.WriteTimeSeries(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCloudMonitoringClient_ListTimeSeries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedTimeSeries := []monitoring.TimeSeriesData{
		{
			MetricType:   "custom.googleapis.com/test_metric",
			ResourceType: "global",
			Values: []monitoring.MetricValue{
				{
					Value:     42.0,
					Timestamp: time.Now(),
				},
			},
		},
	}

	mockClient := mocks.NewMockMonitoringClientInterface(ctrl)
	client := monitoring.NewWithClient(mockClient, "test-project")

	req := monitoring.ListTimeSeriesRequest{
		Filter: "metric.type=\"custom.googleapis.com/test_metric\"",
	}
	req.Interval.StartTime = time.Now().Add(-1 * time.Hour)
	req.Interval.EndTime = time.Now()

	// Set expectation for ListTimeSeries call
	mockClient.EXPECT().
		ListTimeSeries(gomock.Any(), req).
		Return(expectedTimeSeries, nil).
		Times(1)

	result, err := client.ListTimeSeries(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 time series, got %d", len(result))
	}

	if result[0].MetricType != expectedTimeSeries[0].MetricType {
		t.Errorf("Expected metric type %s, got %s", expectedTimeSeries[0].MetricType, result[0].MetricType)
	}
}

func TestCloudMonitoringClient_ListMetricDescriptors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedDescriptors := []monitoring.MetricDescriptor{
		{
			Type:        "custom.googleapis.com/test_metric",
			MetricKind:  "GAUGE",
			ValueType:   "DOUBLE",
			Description: "Test metric",
			DisplayName: "Test Metric",
		},
	}

	mockClient := mocks.NewMockMonitoringClientInterface(ctrl)
	client := monitoring.NewWithClient(mockClient, "test-project")

	// Set expectation for ListMetricDescriptors call
	mockClient.EXPECT().
		ListMetricDescriptors(gomock.Any(), "").
		Return(expectedDescriptors, nil).
		Times(1)

	result, err := client.ListMetricDescriptors(context.Background(), "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 metric descriptor, got %d", len(result))
	}

	if result[0].Type != expectedDescriptors[0].Type {
		t.Errorf("Expected metric type %s, got %s", expectedDescriptors[0].Type, result[0].Type)
	}
}

func TestCloudMonitoringClient_DeleteMetricDescriptor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMonitoringClientInterface(ctrl)
	client := monitoring.NewWithClient(mockClient, "test-project")

	// Set expectation for DeleteMetricDescriptor call
	mockClient.EXPECT().
		DeleteMetricDescriptor(gomock.Any(), "custom.googleapis.com/test_metric").
		Return(nil).
		Times(1)

	err := client.DeleteMetricDescriptor(context.Background(), "custom.googleapis.com/test_metric")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

