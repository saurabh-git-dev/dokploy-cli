package dokploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIntegration_ProjectEnvironmentFlow exercises the client and project/environment
// helpers together against a single fake Dokploy HTTP server.
func TestIntegration_ProjectEnvironmentFlow(t *testing.T) {
	t.Helper()

	var createdProjectID string
	var createdProjectName string
	var createdEnvID string
	var createdEnvName string
	var createdEnvProjectID string

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
		_ = json.NewEncoder(w).Encode(map[string]any{"projectId": createdProjectID})
	})

	mux.HandleFunc("/api/environment.create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		createdEnvID = "env-integration-1"
		createdEnvName, _ = body["name"].(string)
		createdEnvProjectID, _ = body["projectId"].(string)
		_ = json.NewEncoder(w).Encode(map[string]any{"environmentId": createdEnvID})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client, err := NewClient(ts.URL, "integration-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	ctx := context.Background()

	t.Logf("[integration] creating project 'Integration Project'")
	projID, err := CreateProject(ctx, client, "ignored", "Integration Project")
	if err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	if projID == "" || projID != createdProjectID {
		t.Fatalf("unexpected project ID: got %q, want %q", projID, createdProjectID)
	}
	if createdProjectName != "Integration Project" {
		t.Errorf("createdProjectName = %q, want %q", createdProjectName, "Integration Project")
	}

	t.Logf("[integration] creating environment 'Integration Env' in project %s", projID)
	envID, err := CreateEnvironment(ctx, client, "ignored", "Integration Env", projID)
	if err != nil {
		t.Fatalf("CreateEnvironment error: %v", err)
	}
	if envID == "" || envID != createdEnvID {
		t.Fatalf("unexpected environment ID: got %q, want %q", envID, createdEnvID)
	}
	if createdEnvName != "Integration Env" {
		t.Errorf("createdEnvName = %q, want %q", createdEnvName, "Integration Env")
	}
	if createdEnvProjectID != projID {
		t.Errorf("createdEnvProjectID = %q, want %q", createdEnvProjectID, projID)
	}

	t.Logf("[integration] created project %s and environment %s successfully", projID, envID)
}
