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

// TestE2E_ComposeAppWithDomain creates a compose app with a sample docker file
// and configures a domain for a service inside the docker compose.
// It only runs when built with -tags e2e and when DOKPLOY_URL and
// DOKPLOY_API_KEY are set.
func TestE2E_ComposeAppWithDomain(t *testing.T) {
	client := newLiveClient(t)

	ctx := context.Background()
	suffix := time.Now().UnixNano()

	// Step 1: Create a project and environment
	projectName := fmt.Sprintf("dokploy-cli-compose-test-%d", suffix)
	t.Logf("[live] creating project %q", projectName)
	projectID, envID, err := dokploy.CreateProject(ctx, client, projectName, "", "production")
	if err != nil {
		t.Fatalf("CreateProject (live) error: %v", err)
	}
	if projectID == "" || envID == "" {
		t.Fatalf("CreateProject (live) returned empty project ID or environment ID")
	}
	t.Logf("[live] created project with ID %s and environment ID %s", projectID, envID)

	// Cleanup function
	cleanup := func() {
		t.Logf("[live] cleaning up project %s", projectID)
		if err := dokploy.DeleteProject(ctx, client, projectID); err != nil {
			t.Logf("[live] warning: cleanup DeleteProject error: %v", err)
		}
	}
	defer cleanup()

	// Step 2: Create a compose app with a sample docker-compose file
	// Using nginx as a generic docker web image
	composeContent := `services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
    environment:
      - NGINX_HOST=localhost
      - NGINX_PORT=80
`

	composeName := fmt.Sprintf("test-compose-%d", suffix)
	t.Logf("[live] creating compose app %q", composeName)
	composeID, err := dokploy.CreateOrUpdateCompose(
		ctx,
		client,
		"",            // no ID, creating new
		composeName,
		envID,
		composeContent,
		map[string]string{
			"TEST_ENV": "test-value",
		},
	)
	if err != nil {
		t.Fatalf("CreateOrUpdateCompose (live) error: %v", err)
	}
	if composeID == "" {
		t.Fatalf("CreateOrUpdateCompose (live) returned empty compose ID")
	}
	t.Logf("[live] created compose app with ID %s", composeID)

	// Step 3: Create a domain for the web service
	domainHost := fmt.Sprintf("test-%d.example.com", suffix)
	t.Logf("[live] creating domain %q for service 'web' on port 8080", domainHost)
	domainID, err := dokploy.CreateOrUpdateDomain(
		ctx,
		client,
		"",           // no ID, creating new
		domainHost,
		"/",
		8080,
		"web",
		composeID,
		"none",
		false,
	)
	if err != nil {
		t.Fatalf("CreateOrUpdateDomain (live) error: %v", err)
	}
	if domainID == "" {
		t.Fatalf("CreateOrUpdateDomain (live) returned empty domain ID")
	}
	t.Logf("[live] created domain with ID %s", domainID)

	// Step 4: Verify the compose app exists
	t.Logf("[live] verifying compose app %s exists", composeID)
	composeData, err := dokploy.GetCompose(ctx, client, composeID, "")
	if err != nil {
		t.Fatalf("GetCompose (live) error: %v", err)
	}
	if composeData == nil {
		t.Fatalf("GetCompose (live) returned nil data")
	}
	t.Logf("[live] verified compose app exists")

	// Step 5: Deploy the compose app
	t.Logf("[live] deploying compose app %s", composeID)
	if err := dokploy.DeployCompose(ctx, client, composeID); err != nil {
		t.Fatalf("DeployCompose (live) error: %v", err)
	}
	t.Logf("[live] compose app deployed successfully")

	// Step 6: Clean up domain explicitly
	t.Logf("[live] deleting domain %s", domainID)
	if err := dokploy.DeleteDomain(ctx, client, domainID); err != nil {
		t.Logf("[live] warning: DeleteDomain error: %v", err)
	} else {
		t.Logf("[live] domain deleted successfully")
	}

	// Step 7: Clean up compose app explicitly
	t.Logf("[live] deleting compose app %s", composeID)
	if err := dokploy.DeleteCompose(ctx, client, composeID, true); err != nil {
		t.Logf("[live] warning: DeleteCompose error: %v", err)
	} else {
		t.Logf("[live] compose app deleted successfully")
	}

	t.Logf("[live] test completed successfully - project %s, compose %s, domain %s", projectID, composeID, domainID)
}
