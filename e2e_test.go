//go:build e2e
// +build e2e

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/saurabh-git-dev/dokploy-cli/dokploy"
)

// loadEnvFromFile loads simple KEY=VALUE pairs from the given .env-style file.
// Lines starting with # or blank lines are ignored.
func loadEnvFromFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}
		// Only set if not already present in the environment.
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

// newLiveClient constructs a Client for a real Dokploy server using
// environment variables. Tests are skipped if configuration is missing.
func newLiveClient(t *testing.T) *dokploy.Client {
	t.Helper()

	// Allow configuring tests via a .env file as well as real env vars.
	// Try current directory and repo root.
	loadEnvFromFile(".env")
	loadEnvFromFile("../.env")

	baseURL := os.Getenv("DOKPLOY_URL")
	apiKey := os.Getenv("DOKPLOY_API_KEY")
	if baseURL == "" || apiKey == "" {
		t.Skip("DOKPLOY_URL or DOKPLOY_API_KEY not set; skipping live Dokploy tests")
	}

	t.Logf("[live] using DOKPLOY_URL=%s", baseURL)

	client, err := dokploy.NewClient(baseURL, apiKey)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	return client
}

// TestE2E_ProjectAndEnvironment performs a small end-to-end flow against a
// real Dokploy server: create project, then clean up.
// It only runs when built with -tags e2e and when DOKPLOY_URL and
// DOKPLOY_API_KEY are set.
func TestE2E_ProjectAndEnvironment(t *testing.T) {
	client := newLiveClient(t)

	ctx := context.Background()
	suffix := time.Now().UnixNano()

	projectName := fmt.Sprintf("dokploy-cli-test-project-%d", suffix)
	t.Logf("[live] creating project %q", projectName)
	projectID, envID, err := dokploy.CreateProject(ctx, client, projectName, "", "production")
	if err != nil {
		t.Fatalf("CreateProject (live) error: %v", err)
	}
	if projectID == "" {
		t.Fatalf("CreateProject (live) returned empty project ID")
	}
	t.Logf("[live] created project with ID %s and environment ID %s", projectID, envID)

	t.Logf("[live] getting project %q", projectName)
	projectID, envID, err = dokploy.GetProject(ctx, client, projectName, "production")
	if err != nil {
		t.Fatalf("GetProject (live) error: %v", err)
	}
	if projectID == "" {
		t.Fatalf("GetProject (live) returned empty project ID")
	}
	t.Logf("[live] Found project with ID %s and environment ID %s", projectID, envID)

	t.Logf("[live] deleting project %s", projectID)
	if err := dokploy.DeleteProject(ctx, client, projectID); err != nil {
		t.Fatalf("DeleteProject (live) error: %v", err)
	}
	t.Logf("[live] cleanup complete for project %s and environment %s", projectID, envID)
}
