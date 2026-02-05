package dokploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEnvironment_CallsEnvironmentCreate(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"environmentId": "env-123"})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	id, err := CreateEnvironment(context.Background(), client, "ignored", "staging", "proj-1")
	if err != nil {
		t.Fatalf("CreateEnvironment error: %v", err)
	}
	if gotPath != "/api/environment.create" {
		t.Errorf("path = %q, want %q", gotPath, "/api/environment.create")
	}
	if gotBody["name"] != "staging" {
		t.Errorf("name = %v, want %v", gotBody["name"], "staging")
	}
	if gotBody["projectId"] != "proj-1" {
		t.Errorf("projectId = %v, want %v", gotBody["projectId"], "proj-1")
	}
	if id != "env-123" {
		t.Errorf("id = %q, want %q", id, "env-123")
	}
}

func TestDeleteEnvironment_CallsEnvironmentRemove(t *testing.T) {
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

	if err := DeleteEnvironment(context.Background(), client, "env-123"); err != nil {
		t.Fatalf("DeleteEnvironment error: %v", err)
	}
	if gotPath != "/api/environment.remove" {
		t.Errorf("path = %q, want %q", gotPath, "/api/environment.remove")
	}
	if gotBody["environmentId"] != "env-123" {
		t.Errorf("environmentId = %v, want %v", gotBody["environmentId"], "env-123")
	}
}
