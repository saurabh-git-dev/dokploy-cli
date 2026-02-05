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
		_ = json.NewEncoder(w).Encode(map[string]any{"projectId": "proj-123"})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	id, err := CreateProject(context.Background(), client, "ignored", "My Project")
	if err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	if gotPath != "/api/project.create" {
		t.Errorf("path = %q, want %q", gotPath, "/api/project.create")
	}
	if gotBody["name"] != "My Project" {
		t.Errorf("name = %v, want %v", gotBody["name"], "My Project")
	}
	if id != "proj-123" {
		t.Errorf("id = %q, want %q", id, "proj-123")
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
