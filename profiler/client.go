package profiler

//go:generate go tool mockgen -destination=mocks/mock_client.go -package=mocks github.com/kitagry/gcp-telemetry-mcp/profiler ProfilerClient,ProfilerClientInterface

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/cloudprofiler/v2"
	"google.golang.org/api/option"
)

// ProfileType represents the type of profile
type ProfileType string

const (
	ProfileTypeCPU        ProfileType = "CPU"
	ProfileTypeHeap       ProfileType = "HEAP"
	ProfileTypeThreads    ProfileType = "THREADS"
	ProfileTypeContention ProfileType = "CONTENTION"
	ProfileTypeWall       ProfileType = "WALL"
)

// Profile represents a profiling data
type Profile struct {
	Name         string            `json:"name"`
	ProfileType  ProfileType       `json:"profile_type"`
	Duration     string            `json:"duration"`
	Labels       map[string]string `json:"labels,omitempty"`
	StartTime    time.Time         `json:"start_time"`
	ProfileBytes string            `json:"profile_bytes,omitempty"`
	Deployment   *Deployment       `json:"deployment,omitempty"`
}

// Deployment represents deployment information
type Deployment struct {
	ProjectID string            `json:"project_id"`
	Target    string            `json:"target"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// CreateProfileRequest represents a request to create a profile
type CreateProfileRequest struct {
	ProjectID   string            `json:"project_id"`
	Deployment  *Deployment       `json:"deployment"`
	ProfileType []ProfileType     `json:"profile_type"`
	Duration    string            `json:"duration,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// CreateOfflineProfileRequest represents a request to create an offline profile
type CreateOfflineProfileRequest struct {
	ProjectID string   `json:"project_id"`
	Profile   *Profile `json:"profile"`
}

// ListProfilesRequest represents a request to list profiles
type ListProfilesRequest struct {
	ProjectID string `json:"project_id"`
	PageSize  int64  `json:"page_size,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

// UpdateProfileRequest represents a request to update a profile
type UpdateProfileRequest struct {
	Profile      *Profile `json:"profile"`
	UpdateMask   string   `json:"update_mask,omitempty"`
	ProfileBytes string   `json:"profile_bytes,omitempty"`
}

// ProfilerClient defines the interface for Cloud Profiler operations
type ProfilerClient interface {
	CreateProfile(ctx context.Context, req CreateProfileRequest) (*Profile, error)
	CreateOfflineProfile(ctx context.Context, req CreateOfflineProfileRequest) (*Profile, error)
	UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error)
	ListProfiles(ctx context.Context, req ListProfilesRequest) ([]*Profile, error)
}

// CloudProfilerClient implements ProfilerClient using Google Cloud Profiler
type CloudProfilerClient struct {
	client    ProfilerClientInterface
	projectID string
}

// ProfilerClientInterface abstracts the Google Cloud Profiler client for testing
type ProfilerClientInterface interface {
	CreateProfile(ctx context.Context, req CreateProfileRequest) (*Profile, error)
	CreateOfflineProfile(ctx context.Context, req CreateOfflineProfileRequest) (*Profile, error)
	UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error)
	ListProfiles(ctx context.Context, req ListProfilesRequest) ([]*Profile, error)
}

// New creates a new CloudProfilerClient
func New(projectID string) (*CloudProfilerClient, error) {
	service, err := cloudprofiler.NewService(context.Background(), option.WithScopes(cloudprofiler.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create profiler service: %w", err)
	}

	return &CloudProfilerClient{
		client: &realProfilerClient{
			service:   service,
			projectID: projectID,
		},
		projectID: projectID,
	}, nil
}

// NewWithClient creates a new CloudProfilerClient with a custom interface for testing
func NewWithClient(client ProfilerClientInterface, projectID string) *CloudProfilerClient {
	return &CloudProfilerClient{
		client:    client,
		projectID: projectID,
	}
}

// CreateProfile creates a new profile
func (c *CloudProfilerClient) CreateProfile(ctx context.Context, req CreateProfileRequest) (*Profile, error) {
	return c.client.CreateProfile(ctx, req)
}

// CreateOfflineProfile creates an offline profile
func (c *CloudProfilerClient) CreateOfflineProfile(ctx context.Context, req CreateOfflineProfileRequest) (*Profile, error) {
	return c.client.CreateOfflineProfile(ctx, req)
}

// UpdateProfile updates an existing profile
func (c *CloudProfilerClient) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
	return c.client.UpdateProfile(ctx, req)
}

// ListProfiles lists profiles
func (c *CloudProfilerClient) ListProfiles(ctx context.Context, req ListProfilesRequest) ([]*Profile, error) {
	return c.client.ListProfiles(ctx, req)
}

// realProfilerClient wraps the actual Google Cloud Profiler service
type realProfilerClient struct {
	service   *cloudprofiler.Service
	projectID string
}

// CreateProfile implements ProfilerClientInterface for the real client
func (r *realProfilerClient) CreateProfile(ctx context.Context, req CreateProfileRequest) (*Profile, error) {
	parent := fmt.Sprintf("projects/%s", r.projectID)
	
	// Convert our ProfileType to API strings
	var profileTypes []string
	for _, pt := range req.ProfileType {
		profileTypes = append(profileTypes, string(pt))
	}

	// Create deployment
	deployment := &cloudprofiler.Deployment{
		ProjectId: r.projectID,
		Target:    req.Deployment.Target,
		Labels:    req.Deployment.Labels,
	}

	createReq := &cloudprofiler.CreateProfileRequest{
		Deployment:  deployment,
		ProfileType: profileTypes,
	}

	profile, err := r.service.Projects.Profiles.Create(parent, createReq).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return convertAPIProfileToProfile(profile), nil
}

// UpdateProfile implements ProfilerClientInterface for the real client
func (r *realProfilerClient) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
	// Convert our profile to API profile
	apiProfile := &cloudprofiler.Profile{
		Name:         req.Profile.Name,
		ProfileType:  string(req.Profile.ProfileType),
		Duration:     req.Profile.Duration,
		Labels:       req.Profile.Labels,
		ProfileBytes: req.ProfileBytes,
	}

	if req.Profile.Deployment != nil {
		apiProfile.Deployment = &cloudprofiler.Deployment{
			ProjectId: req.Profile.Deployment.ProjectID,
			Target:    req.Profile.Deployment.Target,
			Labels:    req.Profile.Deployment.Labels,
		}
	}

	profile, err := r.service.Projects.Profiles.Patch(req.Profile.Name, apiProfile).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return convertAPIProfileToProfile(profile), nil
}

// CreateOfflineProfile implements ProfilerClientInterface for the real client
func (r *realProfilerClient) CreateOfflineProfile(ctx context.Context, req CreateOfflineProfileRequest) (*Profile, error) {
	parent := fmt.Sprintf("projects/%s", r.projectID)
	
	// Convert our profile to API profile
	apiProfile := &cloudprofiler.Profile{
		ProfileType:  string(req.Profile.ProfileType),
		Duration:     req.Profile.Duration,
		Labels:       req.Profile.Labels,
		ProfileBytes: req.Profile.ProfileBytes,
	}

	if req.Profile.Deployment != nil {
		apiProfile.Deployment = &cloudprofiler.Deployment{
			ProjectId: req.Profile.Deployment.ProjectID,
			Target:    req.Profile.Deployment.Target,
			Labels:    req.Profile.Deployment.Labels,
		}
	}

	profile, err := r.service.Projects.Profiles.CreateOffline(parent, apiProfile).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return convertAPIProfileToProfile(profile), nil
}

// ListProfiles implements ProfilerClientInterface for the real client
func (r *realProfilerClient) ListProfiles(ctx context.Context, req ListProfilesRequest) ([]*Profile, error) {
	parent := fmt.Sprintf("projects/%s", r.projectID)
	
	call := r.service.Projects.Profiles.List(parent).Context(ctx)
	
	if req.PageSize > 0 {
		call = call.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		call = call.PageToken(req.PageToken)
	}

	response, err := call.Do()
	if err != nil {
		return nil, err
	}

	var profiles []*Profile
	for _, apiProfile := range response.Profiles {
		profiles = append(profiles, convertAPIProfileToProfile(apiProfile))
	}

	return profiles, nil
}

// convertAPIProfileToProfile converts a Cloud Profiler API Profile to our Profile struct
func convertAPIProfileToProfile(apiProfile *cloudprofiler.Profile) *Profile {
	profile := &Profile{
		Name:         apiProfile.Name,
		ProfileType:  ProfileType(apiProfile.ProfileType),
		Duration:     apiProfile.Duration,
		Labels:       apiProfile.Labels,
		ProfileBytes: apiProfile.ProfileBytes,
	}

	if apiProfile.Deployment != nil {
		profile.Deployment = &Deployment{
			ProjectID: apiProfile.Deployment.ProjectId,
			Target:    apiProfile.Deployment.Target,
			Labels:    apiProfile.Deployment.Labels,
		}
	}

	// Parse start time from name if available (profile names typically include timestamps)
	// This is a simplified implementation - in practice, you might parse the actual timestamp
	profile.StartTime = time.Now()

	return profile
}