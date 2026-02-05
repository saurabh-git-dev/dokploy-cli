package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saurabh-git-dev/dokploy-cli/dokploy"
)

// TestIntegration_ProjectEnvironmentFlow exercises the client and project/environment
// helpers together against a single fake Dokploy HTTP server.
func TestIntegration_ProjectEnvironmentFlow(t *testing.T) {
	t.Helper()

	var createdProjectID string
	var createdProjectName string

	mux := http.NewServeMux()

	mux.HandleFunc("/api/project.create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		createdProjectID = "proj-integration-1"
		if name, _ := body["name"].(string); name != "Integration Project" {
			t.Errorf("project name = %q, want %q", name, "Integration Project")
		}
		createdProjectName, _ = body["name"].(string)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"project": map[string]any{
				"projectId":      createdProjectID,
				"name":           createdProjectName,
				"description":    nil,
				"createdAt":      "2026-02-05T09:27:24.786Z",
				"organizationId": "org-integration",
				"env":            "",
			},
			"environment": map[string]any{
				"environmentId": "env-integration-default",
				"name":          "production",
				"description":   "Default environment",
				"createdAt":     "2026-02-05T09:27:24.790Z",
				"env":           "",
				"projectId":     createdProjectID,
			},
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client, err := dokploy.NewClient(ts.URL, "integration-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	ctx := context.Background()

	t.Logf("[integration] creating project 'Integration Project'")
	projID, envID, err := dokploy.CreateProject(ctx, client, "Integration Project", "", "production")
	if err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	if projID == "" || projID != createdProjectID {
		t.Fatalf("unexpected project ID: got %q, want %q", projID, createdProjectID)
	}
	if createdProjectName != "Integration Project" {
		t.Errorf("createdProjectName = %q, want %q", createdProjectName, "Integration Project")
	}

	t.Logf("[integration] created project %s and environment %s successfully", projID, envID)
}
