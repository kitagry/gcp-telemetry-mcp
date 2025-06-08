package monitoring

import (
	"context"
	"testing"
	"time"
)

// mockMonitoringClient implements MonitoringClientInterface for testing
type mockMonitoringClient struct {
	createMetricDescriptorCalled   bool
	writeTimeSeriesCalled          bool
	listTimeSeriesCalled           bool
	listMetricDescriptorsCalled    bool
	deleteMetricDescriptorCalled   bool
	
	createMetricDescriptorError    error
	writeTimeSeriesError           error
	listTimeSeriesError            error
	listMetricDescriptorsError     error
	deleteMetricDescriptorError    error
	
	timeSeriesResponse             []TimeSeriesData
	metricDescriptorsResponse      []MetricDescriptor
}

func (m *mockMonitoringClient) CreateMetricDescriptor(ctx context.Context, req CreateMetricRequest) error {
	m.createMetricDescriptorCalled = true
	return m.createMetricDescriptorError
}

func (m *mockMonitoringClient) WriteTimeSeries(ctx context.Context, req WriteTimeSeriesRequest) error {
	m.writeTimeSeriesCalled = true
	return m.writeTimeSeriesError
}

func (m *mockMonitoringClient) ListTimeSeries(ctx context.Context, req ListTimeSeriesRequest) ([]TimeSeriesData, error) {
	m.listTimeSeriesCalled = true
	return m.timeSeriesResponse, m.listTimeSeriesError
}

func (m *mockMonitoringClient) ListMetricDescriptors(ctx context.Context, filter string) ([]MetricDescriptor, error) {
	m.listMetricDescriptorsCalled = true
	return m.metricDescriptorsResponse, m.listMetricDescriptorsError
}

func (m *mockMonitoringClient) DeleteMetricDescriptor(ctx context.Context, metricType string) error {
	m.deleteMetricDescriptorCalled = true
	return m.deleteMetricDescriptorError
}

func TestCloudMonitoringClient_CreateMetricDescriptor(t *testing.T) {
	mockClient := &mockMonitoringClient{}
	client := &CloudMonitoringClient{
		client:    mockClient,
		projectID: "test-project",
	}

	req := CreateMetricRequest{
		MetricDescriptor: MetricDescriptor{
			Type:        "custom.googleapis.com/test_metric",
			MetricKind:  "GAUGE",
			ValueType:   "DOUBLE",
			Description: "Test metric",
			DisplayName: "Test Metric",
		},
	}

	err := client.CreateMetricDescriptor(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockClient.createMetricDescriptorCalled {
		t.Error("Expected CreateMetricDescriptor to be called")
	}
}

func TestCloudMonitoringClient_WriteTimeSeries(t *testing.T) {
	mockClient := &mockMonitoringClient{}
	client := &CloudMonitoringClient{
		client:    mockClient,
		projectID: "test-project",
	}

	req := WriteTimeSeriesRequest{
		TimeSeries: []TimeSeriesData{
			{
				MetricType:   "custom.googleapis.com/test_metric",
				ResourceType: "global",
				Values: []MetricValue{
					{
						Value:     42.0,
						Timestamp: time.Now(),
					},
				},
			},
		},
	}

	err := client.WriteTimeSeries(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockClient.writeTimeSeriesCalled {
		t.Error("Expected WriteTimeSeries to be called")
	}
}

func TestCloudMonitoringClient_ListTimeSeries(t *testing.T) {
	expectedTimeSeries := []TimeSeriesData{
		{
			MetricType:   "custom.googleapis.com/test_metric",
			ResourceType: "global",
			Values: []MetricValue{
				{
					Value:     42.0,
					Timestamp: time.Now(),
				},
			},
		},
	}

	mockClient := &mockMonitoringClient{
		timeSeriesResponse: expectedTimeSeries,
	}
	client := &CloudMonitoringClient{
		client:    mockClient,
		projectID: "test-project",
	}

	req := ListTimeSeriesRequest{
		Filter: "metric.type=\"custom.googleapis.com/test_metric\"",
	}
	req.Interval.StartTime = time.Now().Add(-1 * time.Hour)
	req.Interval.EndTime = time.Now()

	result, err := client.ListTimeSeries(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockClient.listTimeSeriesCalled {
		t.Error("Expected ListTimeSeries to be called")
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 time series, got %d", len(result))
	}

	if result[0].MetricType != expectedTimeSeries[0].MetricType {
		t.Errorf("Expected metric type %s, got %s", expectedTimeSeries[0].MetricType, result[0].MetricType)
	}
}

func TestCloudMonitoringClient_ListMetricDescriptors(t *testing.T) {
	expectedDescriptors := []MetricDescriptor{
		{
			Type:        "custom.googleapis.com/test_metric",
			MetricKind:  "GAUGE",
			ValueType:   "DOUBLE",
			Description: "Test metric",
			DisplayName: "Test Metric",
		},
	}

	mockClient := &mockMonitoringClient{
		metricDescriptorsResponse: expectedDescriptors,
	}
	client := &CloudMonitoringClient{
		client:    mockClient,
		projectID: "test-project",
	}

	result, err := client.ListMetricDescriptors(context.Background(), "")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockClient.listMetricDescriptorsCalled {
		t.Error("Expected ListMetricDescriptors to be called")
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 metric descriptor, got %d", len(result))
	}

	if result[0].Type != expectedDescriptors[0].Type {
		t.Errorf("Expected metric type %s, got %s", expectedDescriptors[0].Type, result[0].Type)
	}
}

func TestCloudMonitoringClient_DeleteMetricDescriptor(t *testing.T) {
	mockClient := &mockMonitoringClient{}
	client := &CloudMonitoringClient{
		client:    mockClient,
		projectID: "test-project",
	}

	err := client.DeleteMetricDescriptor(context.Background(), "custom.googleapis.com/test_metric")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mockClient.deleteMetricDescriptorCalled {
		t.Error("Expected DeleteMetricDescriptor to be called")
	}
}