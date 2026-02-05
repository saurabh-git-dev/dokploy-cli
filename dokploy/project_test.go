package dokploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject_CallsProjectCreate(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"project": map[string]any{
				"projectId":      "proj-123",
				"name":           "My Project",
				"description":    nil,
				"createdAt":      "2026-02-05T09:27:24.786Z",
				"organizationId": "org-1",
				"env":            "",
			},
			"environment": map[string]any{
				"environmentId": "env-123",
				"name":          "production",
				"description":   "Production environment",
				"createdAt":     "2026-02-05T09:27:24.790Z",
				"env":           "",
				"projectId":     "proj-123",
			},
		})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	projId, _, err := CreateProject(context.Background(), client, "My Project", "", "production")
	if err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	if gotPath != "/api/project.create" {
		t.Errorf("path = %q, want %q", gotPath, "/api/project.create")
	}
	if gotBody["name"] != "My Project" {
		t.Errorf("name = %v, want %v", gotBody["name"], "My Project")
	}
	if projId != "proj-123" {
		t.Errorf("id = %q, want %q", projId, "proj-123")
	}
}

func TestDeleteProject_CallsProjectRemove(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if err := DeleteProject(context.Background(), client, "proj-123"); err != nil {
		t.Fatalf("DeleteProject error: %v", err)
	}
	if gotPath != "/api/project.remove" {
		t.Errorf("path = %q, want %q", gotPath, "/api/project.remove")
	}
	if gotBody["projectId"] != "proj-123" {
		t.Errorf("projectId = %v, want %v", gotBody["projectId"], "proj-123")
	}
}

func TestListProjects_CallsProjectAll(t *testing.T) {
	t.Helper()

	wanted := []Project{
		{
			ProjectID:   "proj-1",
			Name:        "project-one",
			Description: "first",
			Env:         "production",
			Environments: []ProjectEnvironment{
				{EnvironmentID: "env-1", Name: "production", Description: "prod"},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project.all" {
			t.Fatalf("expected path /api/project.all, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("expected method GET, got %s", r.Method)
		}
		if err := json.NewEncoder(w).Encode(wanted); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	ctx := context.Background()

	projects, err := ListProjects(ctx, client)
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ProjectID != "proj-1" || projects[0].Name != "project-one" {
		t.Fatalf("unexpected project: %+v", projects[0])
	}
}

func TestGetProject_FindsProjectAndEnvironment(t *testing.T) {
	t.Helper()

	response := []Project{
		{
			ProjectID:   "proj-1",
			Name:        "project-one",
			Description: "first",
			Env:         "production",
			Environments: []ProjectEnvironment{
				{EnvironmentID: "env-1", Name: "production", Description: "prod"},
				{EnvironmentID: "env-2", Name: "staging", Description: "stg"},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project.all" {
			t.Fatalf("expected path /api/project.all, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("expected method GET, got %s", r.Method)
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	ctx := context.Background()

	projectID, environmentID, err := GetProject(ctx, client, "project-one", "staging")
	if err != nil {
		t.Fatalf("GetProject returned error: %v", err)
	}
	if projectID != "proj-1" {
		t.Fatalf("expected projectID proj-1, got %s", projectID)
	}
	if environmentID != "env-2" {
		t.Fatalf("expected environmentID env-2, got %s", environmentID)
	}
}

func TestGetProject_ProjectNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode([]Project{}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	ctx := context.Background()

	_, _, err = GetProject(ctx, client, "missing", "production")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetProject_EnvironmentNotFound(t *testing.T) {
	response := []Project{
		{
			ProjectID:   "proj-1",
			Name:        "project-one",
			Description: "first",
			Env:         "production",
			Environments: []ProjectEnvironment{
				{EnvironmentID: "env-1", Name: "production", Description: "prod"},
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	ctx := context.Background()

	_, _, err = GetProject(ctx, client, "project-one", "staging")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
