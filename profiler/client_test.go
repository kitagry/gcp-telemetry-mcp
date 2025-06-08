package profiler_test

import (
	"context"
	"testing"
	"time"

	"github.com/kitagry/gcp-telemetry-mcp/profiler"
	"github.com/kitagry/gcp-telemetry-mcp/profiler/mocks"
	"go.uber.org/mock/gomock"
)

func TestCloudProfilerClient_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedProfile := &profiler.Profile{
		Name:        "projects/test-project/profiles/profile123",
		ProfileType: profiler.ProfileTypeCPU,
		Duration:    "60s",
		Labels: map[string]string{
			"service": "test-service",
		},
		StartTime: time.Now(),
		Deployment: &profiler.Deployment{
			ProjectID: "test-project",
			Target:    "test-target",
			Labels: map[string]string{
				"version": "v1.0.0",
			},
		},
	}

	mockClient := mocks.NewMockProfilerClientInterface(ctrl)
	client := profiler.NewWithClient(mockClient, "test-project")

	req := profiler.CreateProfileRequest{
		ProjectID: "test-project",
		Deployment: &profiler.Deployment{
			ProjectID: "test-project",
			Target:    "test-target",
			Labels: map[string]string{
				"version": "v1.0.0",
			},
		},
		ProfileType: []profiler.ProfileType{profiler.ProfileTypeCPU},
		Duration:    "60s",
		Labels: map[string]string{
			"service": "test-service",
		},
	}

	// Set expectation for CreateProfile call
	mockClient.EXPECT().
		CreateProfile(gomock.Any(), req).
		Return(expectedProfile, nil).
		Times(1)

	result, err := client.CreateProfile(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Name != expectedProfile.Name {
		t.Errorf("Expected profile name %s, got %s", expectedProfile.Name, result.Name)
	}

	if result.ProfileType != expectedProfile.ProfileType {
		t.Errorf("Expected profile type %s, got %s", expectedProfile.ProfileType, result.ProfileType)
	}

	if result.Duration != expectedProfile.Duration {
		t.Errorf("Expected duration %s, got %s", expectedProfile.Duration, result.Duration)
	}
}

func TestCloudProfilerClient_CreateOfflineProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedProfile := &profiler.Profile{
		Name:         "projects/test-project/profiles/offline123",
		ProfileType:  profiler.ProfileTypeHeap,
		Duration:     "30s",
		ProfileBytes: "base64encodeddata",
		Labels: map[string]string{
			"service": "test-service",
		},
		StartTime: time.Now(),
	}

	mockClient := mocks.NewMockProfilerClientInterface(ctrl)
	client := profiler.NewWithClient(mockClient, "test-project")

	req := profiler.CreateOfflineProfileRequest{
		ProjectID: "test-project",
		Profile: &profiler.Profile{
			ProfileType:  profiler.ProfileTypeHeap,
			Duration:     "30s",
			ProfileBytes: "base64encodeddata",
			Labels: map[string]string{
				"service": "test-service",
			},
		},
	}

	// Set expectation for CreateOfflineProfile call
	mockClient.EXPECT().
		CreateOfflineProfile(gomock.Any(), req).
		Return(expectedProfile, nil).
		Times(1)

	result, err := client.CreateOfflineProfile(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Name != expectedProfile.Name {
		t.Errorf("Expected profile name %s, got %s", expectedProfile.Name, result.Name)
	}

	if result.ProfileType != expectedProfile.ProfileType {
		t.Errorf("Expected profile type %s, got %s", expectedProfile.ProfileType, result.ProfileType)
	}

	if result.ProfileBytes != expectedProfile.ProfileBytes {
		t.Errorf("Expected profile bytes %s, got %s", expectedProfile.ProfileBytes, result.ProfileBytes)
	}
}

func TestCloudProfilerClient_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedProfile := &profiler.Profile{
		Name:         "projects/test-project/profiles/profile123",
		ProfileType:  profiler.ProfileTypeCPU,
		Duration:     "60s",
		ProfileBytes: "updatedbase64data",
		Labels: map[string]string{
			"service": "updated-service",
		},
		StartTime: time.Now(),
	}

	mockClient := mocks.NewMockProfilerClientInterface(ctrl)
	client := profiler.NewWithClient(mockClient, "test-project")

	req := profiler.UpdateProfileRequest{
		Profile: &profiler.Profile{
			Name:        "projects/test-project/profiles/profile123",
			ProfileType: profiler.ProfileTypeCPU,
			Duration:    "60s",
			Labels: map[string]string{
				"service": "updated-service",
			},
		},
		ProfileBytes: "updatedbase64data",
		UpdateMask:   "labels,profile_bytes",
	}

	// Set expectation for UpdateProfile call
	mockClient.EXPECT().
		UpdateProfile(gomock.Any(), req).
		Return(expectedProfile, nil).
		Times(1)

	result, err := client.UpdateProfile(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Name != expectedProfile.Name {
		t.Errorf("Expected profile name %s, got %s", expectedProfile.Name, result.Name)
	}

	if result.ProfileBytes != expectedProfile.ProfileBytes {
		t.Errorf("Expected profile bytes %s, got %s", expectedProfile.ProfileBytes, result.ProfileBytes)
	}

	if result.Labels["service"] != expectedProfile.Labels["service"] {
		t.Errorf("Expected service label %s, got %s", expectedProfile.Labels["service"], result.Labels["service"])
	}
}

func TestCloudProfilerClient_ListProfiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedProfiles := []*profiler.Profile{
		{
			Name:        "projects/test-project/profiles/profile1",
			ProfileType: profiler.ProfileTypeCPU,
			Duration:    "60s",
			StartTime:   time.Now().Add(-1 * time.Hour),
		},
		{
			Name:        "projects/test-project/profiles/profile2",
			ProfileType: profiler.ProfileTypeHeap,
			Duration:    "30s",
			StartTime:   time.Now().Add(-30 * time.Minute),
		},
	}

	mockClient := mocks.NewMockProfilerClientInterface(ctrl)
	client := profiler.NewWithClient(mockClient, "test-project")

	req := profiler.ListProfilesRequest{
		ProjectID: "test-project",
		PageSize:  50,
	}

	// Set expectation for ListProfiles call
	mockClient.EXPECT().
		ListProfiles(gomock.Any(), req).
		Return(expectedProfiles, nil).
		Times(1)

	result, err := client.ListProfiles(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(result))
	}

	if result[0].Name != expectedProfiles[0].Name {
		t.Errorf("Expected profile name %s, got %s", expectedProfiles[0].Name, result[0].Name)
	}

	if result[1].ProfileType != expectedProfiles[1].ProfileType {
		t.Errorf("Expected profile type %s, got %s", expectedProfiles[1].ProfileType, result[1].ProfileType)
	}
}