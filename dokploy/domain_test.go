package dokploy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateDomain_WhenNoneExists_CallsDomainCreate(t *testing.T) {
	t.Helper()

	var gotPaths []string
	var gotCreateBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)

		switch r.URL.Path {
		case "/api/domain.byComposeId":
			// return empty list -> no existing domain
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]domainByComposeItem{})
		case "/api/domain.create":
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"domainId": "dom-123"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	id, err := CreateOrUpdateDomain(context.Background(), client, "", "example.com", "/", 80, "web", "cmp-1", "none", true)
	if err != nil {
		t.Fatalf("CreateOrUpdateDomain error: %v", err)
	}

	if len(gotPaths) != 2 {
		t.Fatalf("expected 2 calls (byComposeId, create), got %d", len(gotPaths))
	}
	if gotPaths[0] != "/api/domain.byComposeId" || gotPaths[1] != "/api/domain.create" {
		t.Errorf("paths = %v, want [/api/domain.byComposeId /api/domain.create]", gotPaths)
	}
	if gotCreateBody["host"] != "example.com" {
		t.Errorf("host = %v, want %v", gotCreateBody["host"], "example.com")
	}
	if gotCreateBody["composeId"] != "cmp-1" {
		t.Errorf("composeId = %v, want %v", gotCreateBody["composeId"], "cmp-1")
	}
	if id != "dom-123" {
		t.Errorf("id = %q, want %q", id, "dom-123")
	}
}

func TestCreateDomain_WhenExists_UsesUpdate(t *testing.T) {
	t.Helper()

	var gotPaths []string
	var gotUpdateBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)

		switch r.URL.Path {
		case "/api/domain.byComposeId":
			items := []domainByComposeItem{{DomainID: "dom-1", Host: "example.com", Path: "/"}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(items)
		case "/api/domain.update":
			if err := json.NewDecoder(r.Body).Decode(&gotUpdateBody); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"domainId": "dom-1"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	id, err := CreateOrUpdateDomain(context.Background(), client, "", "example.com", "/", 80, "web", "cmp-1", "none", true)
	if err != nil {
		t.Fatalf("CreateOrUpdateDomain error: %v", err)
	}

	if len(gotPaths) != 2 {
		t.Fatalf("expected 2 calls (byComposeId, update), got %d", len(gotPaths))
	}
	if gotPaths[0] != "/api/domain.byComposeId" || gotPaths[1] != "/api/domain.update" {
		t.Errorf("paths = %v, want [/api/domain.byComposeId /api/domain.update]", gotPaths)
	}
	if gotUpdateBody["domainId"] != "dom-1" {
		t.Errorf("domainId = %v, want %v", gotUpdateBody["domainId"], "dom-1")
	}
	if gotUpdateBody["composeId"] != "cmp-1" {
		t.Errorf("composeId = %v, want %v", gotUpdateBody["composeId"], "cmp-1")
	}
	if id != "dom-1" {
		t.Errorf("id = %q, want %q", id, "dom-1")
	}
}

func TestDeleteDomain_CallsDomainDelete(t *testing.T) {
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

	if err := DeleteDomain(context.Background(), client, "dom-1"); err != nil {
		t.Fatalf("DeleteDomain error: %v", err)
	}
	if gotPath != "/api/domain.delete" {
		t.Errorf("path = %q, want %q", gotPath, "/api/domain.delete")
	}
	if gotBody["domainId"] != "dom-1" {
		t.Errorf("domainId = %v, want %v", gotBody["domainId"], "dom-1")
	}
}
