package monitoring

//go:generate go tool mockgen -destination=mocks/mock_client.go -package=mocks github.com/kitagry/gcp-telemetry-mcp/monitoring MonitoringClient,MonitoringClientInterface

import (
	"context"
	"fmt"
	"maps"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MetricValue represents a metric value with timestamp
type MetricValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// MetricDescriptor represents metadata about a metric
type MetricDescriptor struct {
	Type        string            `json:"type"`
	MetricKind  string            `json:"metric_kind"`
	ValueType   string            `json:"value_type"`
	Description string            `json:"description"`
	DisplayName string            `json:"display_name"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// TimeSeriesData represents time series data for a metric
type TimeSeriesData struct {
	MetricType   string            `json:"metric_type"`
	MetricLabels map[string]string `json:"metric_labels,omitempty"`
	ResourceType string            `json:"resource_type"`
	Values       []MetricValue     `json:"values"`
}

// CreateMetricRequest represents a request to create a custom metric
type CreateMetricRequest struct {
	MetricDescriptor MetricDescriptor `json:"metric_descriptor"`
}

// WriteTimeSeriesRequest represents a request to write time series data
type WriteTimeSeriesRequest struct {
	TimeSeries []TimeSeriesData `json:"time_series"`
}

// ListTimeSeriesRequest represents a request to list time series data
type ListTimeSeriesRequest struct {
	Filter   string `json:"filter"`
	Interval struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	} `json:"interval"`
	Aggregation *AggregationConfig `json:"aggregation,omitempty"`
	PageSize    int                `json:"page_size,omitempty"`
	PageToken   string             `json:"page_token,omitempty"`
}

// AggregationConfig represents aggregation configuration for time series queries
type AggregationConfig struct {
	AlignmentPeriod    string   `json:"alignment_period"`
	PerSeriesAligner   string   `json:"per_series_aligner"`
	CrossSeriesReducer string   `json:"cross_series_reducer,omitempty"`
	GroupByFields      []string `json:"group_by_fields,omitempty"`
}

// ListAvailableMetricsRequest represents a request to list available metrics
type ListAvailableMetricsRequest struct {
	Filter    string `json:"filter,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

// AvailableMetric represents an available metric in Cloud Monitoring
type AvailableMetric struct {
	Type        string        `json:"type"`
	DisplayName string        `json:"display_name"`
	Description string        `json:"description"`
	MetricKind  string        `json:"metric_kind"`
	ValueType   string        `json:"value_type"`
	Unit        string        `json:"unit,omitempty"`
	Labels      []MetricLabel `json:"labels,omitempty"`
	LaunchStage string        `json:"launch_stage,omitempty"`
}

// MetricLabel represents a label for a metric
type MetricLabel struct {
	Key         string `json:"key"`
	ValueType   string `json:"value_type"`
	Description string `json:"description"`
}

// ListTimeSeriesResponse represents a response with time series data and pagination info
type ListTimeSeriesResponse struct {
	TimeSeries    []TimeSeriesData `json:"time_series"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

// MonitoringClient defines the interface for Cloud Monitoring operations
type MonitoringClient interface {
	CreateMetricDescriptor(ctx context.Context, req CreateMetricRequest) error
	WriteTimeSeries(ctx context.Context, req WriteTimeSeriesRequest) error
	ListTimeSeries(ctx context.Context, req ListTimeSeriesRequest) (ListTimeSeriesResponse, error)
	ListMetricDescriptors(ctx context.Context, req ListMetricDescriptorsRequest) (ListMetricDescriptorsResponse, error)
	DeleteMetricDescriptor(ctx context.Context, metricType string) error
	ListAvailableMetrics(ctx context.Context, req ListAvailableMetricsRequest) ([]AvailableMetric, error)
}

// CloudMonitoringClient implements MonitoringClient using Google Cloud Monitoring
type CloudMonitoringClient struct {
	client    MonitoringClientInterface
	projectID string
}

// MonitoringClientInterface abstracts the Google Cloud Monitoring client for testing
type MonitoringClientInterface interface {
	CreateMetricDescriptor(ctx context.Context, req CreateMetricRequest) error
	WriteTimeSeries(ctx context.Context, req WriteTimeSeriesRequest) error
	ListTimeSeries(ctx context.Context, req ListTimeSeriesRequest) (ListTimeSeriesResponse, error)
	ListMetricDescriptors(ctx context.Context, req ListMetricDescriptorsRequest) (ListMetricDescriptorsResponse, error)
	DeleteMetricDescriptor(ctx context.Context, metricType string) error
	ListAvailableMetrics(ctx context.Context, req ListAvailableMetricsRequest) ([]AvailableMetric, error)
}

// New creates a new CloudMonitoringClient
func New(projectID string) (*CloudMonitoringClient, error) {
	metricClient, err := monitoring.NewMetricClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create metric client: %w", err)
	}

	queryClient, err := monitoring.NewQueryClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create query client: %w", err)
	}

	return &CloudMonitoringClient{
		client: &realMonitoringClient{
			metricClient: metricClient,
			queryClient:  queryClient,
			projectID:    projectID,
		},
		projectID: projectID,
	}, nil
}

// NewWithClient creates a new CloudMonitoringClient with a custom interface for testing
func NewWithClient(client MonitoringClientInterface, projectID string) *CloudMonitoringClient {
	return &CloudMonitoringClient{
		client:    client,
		projectID: projectID,
	}
}

// CreateMetricDescriptor creates a custom metric descriptor
func (c *CloudMonitoringClient) CreateMetricDescriptor(ctx context.Context, req CreateMetricRequest) error {
	return c.client.CreateMetricDescriptor(ctx, req)
}

// WriteTimeSeries writes time series data to Cloud Monitoring
func (c *CloudMonitoringClient) WriteTimeSeries(ctx context.Context, req WriteTimeSeriesRequest) error {
	return c.client.WriteTimeSeries(ctx, req)
}

// ListTimeSeries retrieves time series data from Cloud Monitoring
func (c *CloudMonitoringClient) ListTimeSeries(ctx context.Context, req ListTimeSeriesRequest) (ListTimeSeriesResponse, error) {
	return c.client.ListTimeSeries(ctx, req)
}

// ListMetricDescriptors lists metric descriptors with pagination support
func (c *CloudMonitoringClient) ListMetricDescriptors(ctx context.Context, req ListMetricDescriptorsRequest) (ListMetricDescriptorsResponse, error) {
	return c.client.ListMetricDescriptors(ctx, req)
}

// DeleteMetricDescriptor deletes a custom metric descriptor
func (c *CloudMonitoringClient) DeleteMetricDescriptor(ctx context.Context, metricType string) error {
	return c.client.DeleteMetricDescriptor(ctx, metricType)
}

// ListAvailableMetrics lists available metrics in Cloud Monitoring
func (c *CloudMonitoringClient) ListAvailableMetrics(ctx context.Context, req ListAvailableMetricsRequest) ([]AvailableMetric, error) {
	return c.client.ListAvailableMetrics(ctx, req)
}

// realMonitoringClient wraps the actual Google Cloud Monitoring clients
type realMonitoringClient struct {
	metricClient *monitoring.MetricClient
	queryClient  *monitoring.QueryClient
	projectID    string
}

// CreateMetricDescriptor implements MonitoringClientInterface for the real client
func (r *realMonitoringClient) CreateMetricDescriptor(ctx context.Context, req CreateMetricRequest) error {
	// Convert our MetricDescriptor to the protobuf version
	metricKind := metric.MetricDescriptor_GAUGE
	switch req.MetricDescriptor.MetricKind {
	case "GAUGE":
		metricKind = metric.MetricDescriptor_GAUGE
	case "DELTA":
		metricKind = metric.MetricDescriptor_DELTA
	case "CUMULATIVE":
		metricKind = metric.MetricDescriptor_CUMULATIVE
	}

	valueType := metric.MetricDescriptor_DOUBLE
	switch req.MetricDescriptor.ValueType {
	case "BOOL":
		valueType = metric.MetricDescriptor_BOOL
	case "INT64":
		valueType = metric.MetricDescriptor_INT64
	case "DOUBLE":
		valueType = metric.MetricDescriptor_DOUBLE
	case "STRING":
		valueType = metric.MetricDescriptor_STRING
	case "DISTRIBUTION":
		valueType = metric.MetricDescriptor_DISTRIBUTION
	}

	pbReq := &monitoringpb.CreateMetricDescriptorRequest{
		Name: fmt.Sprintf("projects/%s", r.projectID),
		MetricDescriptor: &metric.MetricDescriptor{
			Type:        req.MetricDescriptor.Type,
			MetricKind:  metricKind,
			ValueType:   valueType,
			Description: req.MetricDescriptor.Description,
			DisplayName: req.MetricDescriptor.DisplayName,
		},
	}

	_, err := r.metricClient.CreateMetricDescriptor(ctx, pbReq)
	return err
}

// WriteTimeSeries implements MonitoringClientInterface for the real client
func (r *realMonitoringClient) WriteTimeSeries(ctx context.Context, req WriteTimeSeriesRequest) error {
	var timeSeries []*monitoringpb.TimeSeries

	for _, ts := range req.TimeSeries {
		var points []*monitoringpb.Point
		for _, value := range ts.Values {
			points = append(points, &monitoringpb.Point{
				Interval: &monitoringpb.TimeInterval{
					EndTime: timestamppb.New(value.Timestamp),
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_DoubleValue{
						DoubleValue: value.Value,
					},
				},
			})
		}

		// Convert labels to protobuf format
		metricLabels := make(map[string]string)
		maps.Copy(metricLabels, ts.MetricLabels)

		timeSeries = append(timeSeries, &monitoringpb.TimeSeries{
			Metric: &metric.Metric{
				Type:   ts.MetricType,
				Labels: metricLabels,
			},
			Resource: &monitoredres.MonitoredResource{
				Type: ts.ResourceType,
			},
			Points: points,
		})
	}

	pbReq := &monitoringpb.CreateTimeSeriesRequest{
		Name:       fmt.Sprintf("projects/%s", r.projectID),
		TimeSeries: timeSeries,
	}

	return r.metricClient.CreateTimeSeries(ctx, pbReq)
}

// ListTimeSeries implements MonitoringClientInterface for the real client
func (r *realMonitoringClient) ListTimeSeries(ctx context.Context, req ListTimeSeriesRequest) (ListTimeSeriesResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100 // デフォルトのページサイズ
	}

	pbReq := &monitoringpb.ListTimeSeriesRequest{
		Name:   fmt.Sprintf("projects/%s", r.projectID),
		Filter: req.Filter,
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(req.Interval.StartTime),
			EndTime:   timestamppb.New(req.Interval.EndTime),
		},
		PageSize: int32(pageSize),
	}

	if req.PageToken != "" {
		pbReq.PageToken = req.PageToken
	}

	// Add aggregation if specified
	if req.Aggregation != nil {
		pbReq.Aggregation = &monitoringpb.Aggregation{
			AlignmentPeriod: parseDuration(req.Aggregation.AlignmentPeriod),
		}

		// Set per-series aligner
		switch req.Aggregation.PerSeriesAligner {
		case "ALIGN_MEAN":
			pbReq.Aggregation.PerSeriesAligner = monitoringpb.Aggregation_ALIGN_MEAN
		case "ALIGN_MAX":
			pbReq.Aggregation.PerSeriesAligner = monitoringpb.Aggregation_ALIGN_MAX
		case "ALIGN_MIN":
			pbReq.Aggregation.PerSeriesAligner = monitoringpb.Aggregation_ALIGN_MIN
		case "ALIGN_SUM":
			pbReq.Aggregation.PerSeriesAligner = monitoringpb.Aggregation_ALIGN_SUM
		default:
			pbReq.Aggregation.PerSeriesAligner = monitoringpb.Aggregation_ALIGN_MEAN
		}

		// Set cross-series reducer if specified
		if req.Aggregation.CrossSeriesReducer != "" {
			switch req.Aggregation.CrossSeriesReducer {
			case "REDUCE_MEAN":
				pbReq.Aggregation.CrossSeriesReducer = monitoringpb.Aggregation_REDUCE_MEAN
			case "REDUCE_MAX":
				pbReq.Aggregation.CrossSeriesReducer = monitoringpb.Aggregation_REDUCE_MAX
			case "REDUCE_MIN":
				pbReq.Aggregation.CrossSeriesReducer = monitoringpb.Aggregation_REDUCE_MIN
			case "REDUCE_SUM":
				pbReq.Aggregation.CrossSeriesReducer = monitoringpb.Aggregation_REDUCE_SUM
			}

			pbReq.Aggregation.GroupByFields = req.Aggregation.GroupByFields
		}
	}

	it := r.metricClient.ListTimeSeries(ctx, pbReq)
	var result []TimeSeriesData

	// 指定したページサイズまでデータを取得
	for range pageSize {
		ts, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return ListTimeSeriesResponse{}, err
		}

		var values []MetricValue
		for _, point := range ts.GetPoints() {
			var value float64
			switch v := point.Value.Value.(type) {
			case *monitoringpb.TypedValue_DoubleValue:
				value = v.DoubleValue
			case *monitoringpb.TypedValue_Int64Value:
				value = float64(v.Int64Value)
			case *monitoringpb.TypedValue_BoolValue:
				if v.BoolValue {
					value = 1.0
				} else {
					value = 0.0
				}
			}

			values = append(values, MetricValue{
				Value:     value,
				Timestamp: point.Interval.EndTime.AsTime(),
			})
		}

		result = append(result, TimeSeriesData{
			MetricType:   ts.Metric.Type,
			MetricLabels: ts.Metric.Labels,
			ResourceType: ts.Resource.Type,
			Values:       values,
		})
	}

	// 次のページトークンを取得
	nextPageToken := it.PageInfo().Token

	return ListTimeSeriesResponse{
		TimeSeries:    result,
		NextPageToken: nextPageToken,
	}, nil
}

// ListMetricDescriptorsRequest represents a request to list metric descriptors
type ListMetricDescriptorsRequest struct {
	Filter    string `json:"filter,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

// ListMetricDescriptorsResponse represents a response with metric descriptors and pagination info
type ListMetricDescriptorsResponse struct {
	Descriptors   []MetricDescriptor `json:"descriptors"`
	NextPageToken string             `json:"next_page_token,omitempty"`
}

// ListMetricDescriptors implements MonitoringClientInterface for the real client
func (r *realMonitoringClient) ListMetricDescriptors(ctx context.Context, req ListMetricDescriptorsRequest) (ListMetricDescriptorsResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 5 // default page size
	}

	pbReq := &monitoringpb.ListMetricDescriptorsRequest{
		Name:     fmt.Sprintf("projects/%s", r.projectID),
		Filter:   req.Filter,
		PageSize: int32(pageSize),
	}

	if req.PageToken != "" {
		pbReq.PageToken = req.PageToken
	}

	it := r.metricClient.ListMetricDescriptors(ctx, pbReq)
	var result []MetricDescriptor

	// 指定したページサイズまでデータを取得
	for range pageSize {
		md, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return ListMetricDescriptorsResponse{}, err
		}

		var metricKind string
		switch md.MetricKind {
		case metric.MetricDescriptor_GAUGE:
			metricKind = "GAUGE"
		case metric.MetricDescriptor_DELTA:
			metricKind = "DELTA"
		case metric.MetricDescriptor_CUMULATIVE:
			metricKind = "CUMULATIVE"
		}

		var valueType string
		switch md.ValueType {
		case metric.MetricDescriptor_BOOL:
			valueType = "BOOL"
		case metric.MetricDescriptor_INT64:
			valueType = "INT64"
		case metric.MetricDescriptor_DOUBLE:
			valueType = "DOUBLE"
		case metric.MetricDescriptor_STRING:
			valueType = "STRING"
		case metric.MetricDescriptor_DISTRIBUTION:
			valueType = "DISTRIBUTION"
		}

		result = append(result, MetricDescriptor{
			Type:        md.Type,
			MetricKind:  metricKind,
			ValueType:   valueType,
			Description: md.Description,
			DisplayName: md.DisplayName,
		})
	}

	// 次のページトークンを取得
	nextPageToken := it.PageInfo().Token

	return ListMetricDescriptorsResponse{
		Descriptors:   result,
		NextPageToken: nextPageToken,
	}, nil
}

// DeleteMetricDescriptor implements MonitoringClientInterface for the real client
func (r *realMonitoringClient) DeleteMetricDescriptor(ctx context.Context, metricType string) error {
	pbReq := &monitoringpb.DeleteMetricDescriptorRequest{
		Name: fmt.Sprintf("projects/%s/metricDescriptors/%s", r.projectID, metricType),
	}

	return r.metricClient.DeleteMetricDescriptor(ctx, pbReq)
}

// ListAvailableMetrics implements MonitoringClientInterface for the real client
func (r *realMonitoringClient) ListAvailableMetrics(ctx context.Context, req ListAvailableMetricsRequest) ([]AvailableMetric, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100 // default page size
	}

	pbReq := &monitoringpb.ListMetricDescriptorsRequest{
		Name:     fmt.Sprintf("projects/%s", r.projectID),
		Filter:   req.Filter,
		PageSize: int32(pageSize),
	}

	if req.PageToken != "" {
		pbReq.PageToken = req.PageToken
	}

	it := r.metricClient.ListMetricDescriptors(ctx, pbReq)
	var result []AvailableMetric

	for i := 0; i < pageSize; i++ {
		md, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var metricKind string
		switch md.MetricKind {
		case metric.MetricDescriptor_GAUGE:
			metricKind = "GAUGE"
		case metric.MetricDescriptor_DELTA:
			metricKind = "DELTA"
		case metric.MetricDescriptor_CUMULATIVE:
			metricKind = "CUMULATIVE"
		default:
			metricKind = "GAUGE"
		}

		var valueType string
		switch md.ValueType {
		case metric.MetricDescriptor_BOOL:
			valueType = "BOOL"
		case metric.MetricDescriptor_INT64:
			valueType = "INT64"
		case metric.MetricDescriptor_DOUBLE:
			valueType = "DOUBLE"
		case metric.MetricDescriptor_STRING:
			valueType = "STRING"
		case metric.MetricDescriptor_DISTRIBUTION:
			valueType = "DISTRIBUTION"
		default:
			valueType = "DOUBLE"
		}

		// Convert label descriptors
		var labels []MetricLabel
		for _, labelDesc := range md.Labels {
			labels = append(labels, MetricLabel{
				Key:         labelDesc.Key,
				ValueType:   "STRING", // Default to STRING for now
				Description: labelDesc.Description,
			})
		}

		var launchStage string
		if md.LaunchStage != 0 {
			launchStage = md.LaunchStage.String()
		}

		result = append(result, AvailableMetric{
			Type:        md.Type,
			DisplayName: md.DisplayName,
			Description: md.Description,
			MetricKind:  metricKind,
			ValueType:   valueType,
			Unit:        md.Unit,
			Labels:      labels,
			LaunchStage: launchStage,
		})
	}

	return result, nil
}

// parseDuration converts a duration string to a protobuf duration
func parseDuration(s string) *durationpb.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		// Default to 60 seconds if parsing fails
		d = 60 * time.Second
	}
	return durationpb.New(d)
}
