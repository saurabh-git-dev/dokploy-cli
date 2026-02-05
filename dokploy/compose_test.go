package dokploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCompose_CallsComposeCreate(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"composeId": "cmp-123"})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	envVars := map[string]string{"A": "1", "B": "2"}
	id, err := CreateOrUpdateCompose(context.Background(), client, "", "my-compose", "env-1", "services: {}", envVars)
	if err != nil {
		t.Fatalf("CreateOrUpdateCompose error: %v", err)
	}
	if gotPath != "/api/compose.create" {
		t.Errorf("path = %q, want %q", gotPath, "/api/compose.create")
	}
	if gotBody["name"] != "my-compose" {
		t.Errorf("name = %v, want %v", gotBody["name"], "my-compose")
	}
	if gotBody["environmentId"] != "env-1" {
		t.Errorf("environmentId = %v, want %v", gotBody["environmentId"], "env-1")
	}
	if gotBody["composeFile"] != "services: {}" {
		t.Errorf("composeFile = %v, want %v", gotBody["composeFile"], "services: {}")
	}
	if gotBody["env"] == "" {
		t.Errorf("env should not be empty")
	}
	if id != "cmp-123" {
		t.Errorf("id = %q, want %q", id, "cmp-123")
	}
}

func TestUpdateCompose_CallsComposeUpdate(t *testing.T) {
	t.Helper()

	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"composeId": "cmp-999"})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	id, err := CreateOrUpdateCompose(context.Background(), client, "cmp-123", "my-compose", "env-1", "services: {}", nil)
	if err != nil {
		t.Fatalf("CreateOrUpdateCompose error: %v", err)
	}
	if gotPath != "/api/compose.update" {
		t.Errorf("path = %q, want %q", gotPath, "/api/compose.update")
	}
	if gotBody["composeId"] != "cmp-123" {
		t.Errorf("composeId = %v, want %v", gotBody["composeId"], "cmp-123")
	}
	if id != "cmp-999" {
		t.Errorf("id = %q, want %q", id, "cmp-999")
	}
}

func TestDeleteCompose_PassesDeleteVolumes(t *testing.T) {
	t.Helper()

	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	if err := DeleteCompose(context.Background(), client, "cmp-1", true); err != nil {
		t.Fatalf("DeleteCompose error: %v", err)
	}
	if gotBody["composeId"] != "cmp-1" {
		t.Errorf("composeId = %v, want %v", gotBody["composeId"], "cmp-1")
	}
	if v, ok := gotBody["deleteVolumes"].(bool); !ok || !v {
		t.Errorf("deleteVolumes = %v, want true", gotBody["deleteVolumes"])
	}
}

func TestDeployCompose_CallsComposeDeploy(t *testing.T) {
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

	if err := DeployCompose(context.Background(), client, "cmp-1"); err != nil {
		t.Fatalf("DeployCompose error: %v", err)
	}
	if gotPath != "/api/compose.deploy" {
		t.Errorf("path = %q, want %q", gotPath, "/api/compose.deploy")
	}
	if gotBody["composeId"] != "cmp-1" {
		t.Errorf("composeId = %v, want %v", gotBody["composeId"], "cmp-1")
	}
}
