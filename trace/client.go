package trace

//go:generate go tool mockgen -destination=mocks/mock_client.go -package=mocks github.com/kitagry/gcp-telemetry-mcp/trace TraceClient,TraceClientInterface

import (
	"context"
	"fmt"
	"time"

	trace "cloud.google.com/go/trace/apiv1"
	"cloud.google.com/go/trace/apiv1/tracepb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Span represents a trace span
type Span struct {
	SpanID    string            `json:"span_id"`
	Name      string            `json:"name"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	ParentID  string            `json:"parent_id,omitempty"`
	Kind      string            `json:"kind,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Trace represents a distributed trace
type Trace struct {
	TraceID   string `json:"trace_id"`
	ProjectID string `json:"project_id"`
	Spans     []Span `json:"spans"`
}

// ListTracesRequest represents a request to list traces
type ListTracesRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Filter    string    `json:"filter,omitempty"`
	OrderBy   string    `json:"order_by,omitempty"`
	PageSize  int       `json:"page_size,omitempty"`
	PageToken string    `json:"page_token,omitempty"`
}

// GetTraceRequest represents a request to get a specific trace
type GetTraceRequest struct {
	TraceID string `json:"trace_id"`
}

// PatchTraceRequest represents a request to update trace spans
type PatchTraceRequest struct {
	TraceID string `json:"trace_id"`
	Spans   []Span `json:"spans"`
}

// TraceClient defines the interface for Cloud Trace operations
type TraceClient interface {
	ListTraces(ctx context.Context, req ListTracesRequest) ([]Trace, error)
	GetTrace(ctx context.Context, req GetTraceRequest) (*Trace, error)
	PatchTraces(ctx context.Context, req PatchTraceRequest) error
}

// CloudTraceClient implements TraceClient using Google Cloud Trace
type CloudTraceClient struct {
	client    TraceClientInterface
	projectID string
}

// TraceClientInterface abstracts the Google Cloud Trace client for testing
type TraceClientInterface interface {
	ListTraces(ctx context.Context, req ListTracesRequest) ([]Trace, error)
	GetTrace(ctx context.Context, req GetTraceRequest) (*Trace, error)
	PatchTraces(ctx context.Context, req PatchTraceRequest) error
}

// New creates a new CloudTraceClient
func New(projectID string) (*CloudTraceClient, error) {
	client, err := trace.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace client: %w", err)
	}

	return &CloudTraceClient{
		client: &realTraceClient{
			client:    client,
			projectID: projectID,
		},
		projectID: projectID,
	}, nil
}

// NewWithClient creates a new CloudTraceClient with a custom interface for testing
func NewWithClient(client TraceClientInterface, projectID string) *CloudTraceClient {
	return &CloudTraceClient{
		client:    client,
		projectID: projectID,
	}
}

// ListTraces lists traces from Cloud Trace
func (c *CloudTraceClient) ListTraces(ctx context.Context, req ListTracesRequest) ([]Trace, error) {
	return c.client.ListTraces(ctx, req)
}

// GetTrace gets a specific trace from Cloud Trace
func (c *CloudTraceClient) GetTrace(ctx context.Context, req GetTraceRequest) (*Trace, error) {
	return c.client.GetTrace(ctx, req)
}

// PatchTraces updates trace spans in Cloud Trace
func (c *CloudTraceClient) PatchTraces(ctx context.Context, req PatchTraceRequest) error {
	return c.client.PatchTraces(ctx, req)
}

// realTraceClient wraps the actual Google Cloud Trace client
type realTraceClient struct {
	client    *trace.Client
	projectID string
}

// ListTraces implements TraceClientInterface for the real client
func (r *realTraceClient) ListTraces(ctx context.Context, req ListTracesRequest) ([]Trace, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100 // default page size
	}

	pbReq := &tracepb.ListTracesRequest{
		ProjectId: r.projectID,
		StartTime: timestamppb.New(req.StartTime),
		EndTime:   timestamppb.New(req.EndTime),
		PageSize:  int32(pageSize),
	}

	if req.Filter != "" {
		pbReq.Filter = req.Filter
	}

	if req.OrderBy != "" {
		pbReq.OrderBy = req.OrderBy
	}

	if req.PageToken != "" {
		pbReq.PageToken = req.PageToken
	}

	it := r.client.ListTraces(ctx, pbReq)
	var result []Trace

	for i := 0; i <= pageSize; i++ {
		traceProto, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		trace := convertProtoToTrace(traceProto, r.projectID)
		result = append(result, trace)
	}

	return result, nil
}

// GetTrace implements TraceClientInterface for the real client
func (r *realTraceClient) GetTrace(ctx context.Context, req GetTraceRequest) (*Trace, error) {
	pbReq := &tracepb.GetTraceRequest{
		ProjectId: r.projectID,
		TraceId:   req.TraceID,
	}

	traceProto, err := r.client.GetTrace(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	trace := convertProtoToTrace(traceProto, r.projectID)
	return &trace, nil
}

// PatchTraces implements TraceClientInterface for the real client
func (r *realTraceClient) PatchTraces(ctx context.Context, req PatchTraceRequest) error {
	// Convert spans to protobuf format
	var pbSpans []*tracepb.TraceSpan
	for _, span := range req.Spans {
		pbSpan := &tracepb.TraceSpan{
			SpanId:    parseSpanID(span.SpanID),
			Name:      span.Name,
			StartTime: timestamppb.New(span.StartTime),
			EndTime:   timestamppb.New(span.EndTime),
			Labels:    span.Labels,
		}

		if span.ParentID != "" {
			pbSpan.ParentSpanId = parseSpanID(span.ParentID)
		}

		// Set span kind
		switch span.Kind {
		case "RPC_SERVER":
			pbSpan.Kind = tracepb.TraceSpan_RPC_SERVER
		case "RPC_CLIENT":
			pbSpan.Kind = tracepb.TraceSpan_RPC_CLIENT
		default:
			pbSpan.Kind = tracepb.TraceSpan_SPAN_KIND_UNSPECIFIED
		}

		pbSpans = append(pbSpans, pbSpan)
	}

	pbReq := &tracepb.PatchTracesRequest{
		ProjectId: r.projectID,
		Traces: &tracepb.Traces{
			Traces: []*tracepb.Trace{
				{
					TraceId: req.TraceID,
					Spans:   pbSpans,
				},
			},
		},
	}

	return r.client.PatchTraces(ctx, pbReq)
}

// convertProtoToTrace converts a protobuf Trace to our Trace struct
func convertProtoToTrace(traceProto *tracepb.Trace, projectID string) Trace {
	var spans []Span
	for _, spanProto := range traceProto.Spans {
		span := Span{
			SpanID:    formatSpanID(spanProto.SpanId),
			Name:      spanProto.Name,
			StartTime: spanProto.StartTime.AsTime(),
			EndTime:   spanProto.EndTime.AsTime(),
			Labels:    spanProto.Labels,
		}

		if spanProto.ParentSpanId != 0 {
			span.ParentID = formatSpanID(spanProto.ParentSpanId)
		}

		// Convert span kind
		switch spanProto.Kind {
		case tracepb.TraceSpan_RPC_SERVER:
			span.Kind = "RPC_SERVER"
		case tracepb.TraceSpan_RPC_CLIENT:
			span.Kind = "RPC_CLIENT"
		default:
			span.Kind = "UNSPECIFIED"
		}

		spans = append(spans, span)
	}

	return Trace{
		TraceID:   traceProto.TraceId,
		ProjectID: projectID,
		Spans:     spans,
	}
}

// parseSpanID converts a string span ID to uint64
func parseSpanID(spanID string) uint64 {
	// This is a simplified implementation
	// In a real implementation, you'd need proper parsing
	// For now, we'll use a hash or conversion
	if spanID == "" {
		return 0
	}
	// Simple hash function for demonstration
	var hash uint64 = 0
	for _, char := range spanID {
		hash = hash*31 + uint64(char)
	}
	return hash
}

// formatSpanID converts a uint64 span ID to string
func formatSpanID(spanID uint64) string {
	if spanID == 0 {
		return ""
	}
	return fmt.Sprintf("%016x", spanID)
}
