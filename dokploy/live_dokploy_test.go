package dokploy

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// newLiveClient constructs a Client for a real Dokploy server using
// environment variables. Tests are skipped if configuration is missing.
func newLiveClient(t *testing.T) *Client {
	t.Helper()

	baseURL := os.Getenv("DOKPLOY_URL")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	if baseURL == "" || apiKey == "" {
		t.Skip("DOKPLOY_URL or DOKPLOY_API_KEY not set; skipping live Dokploy tests")
	}

	client, err := NewClient(baseURL, apiKey)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	return client
}

// TestLive_ProjectAndEnvironment performs a small end-to-end flow against a
// real Dokploy server: create project, create environment, then clean up.
// It only runs when DOKPLOY_URL and DOKPLOY_API_KEY are set.
func TestLive_ProjectAndEnvironment(t *testing.T) {
	client := newLiveClient(t)

	ctx := context.Background()
	suffix := time.Now().UnixNano()

	projectName := fmt.Sprintf("dokploy-cli-test-project-%d", suffix)
	projectID, err := CreateProject(ctx, client, "ignored", projectName)
	if err != nil {
		t.Fatalf("CreateProject (live) error: %v", err)
	}
	if projectID == "" {
		t.Fatalf("CreateProject (live) returned empty project ID")
	}

	envName := fmt.Sprintf("dokploy-cli-test-env-%d", suffix)
	envID, err := CreateEnvironment(ctx, client, "ignored", envName, projectID)
	if err != nil {
		t.Fatalf("CreateEnvironment (live) error: %v", err)
	}
	if envID == "" {
		t.Fatalf("CreateEnvironment (live) returned empty environment ID")
	}

	// Best-effort cleanup; failures here still fail the test so we know
	// if the Dokploy API rejected deletes.
	if err := DeleteEnvironment(ctx, client, envID); err != nil {
		t.Fatalf("DeleteEnvironment (live) error: %v", err)
	}
	if err := DeleteProject(ctx, client, projectID); err != nil {
		t.Fatalf("DeleteProject (live) error: %v", err)
	}
}
