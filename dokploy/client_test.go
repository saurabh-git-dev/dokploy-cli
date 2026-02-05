package dokploy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDo_Success(t *testing.T) {
	t.Helper()

	var gotMethod, gotPath, gotAPIKey string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		gotAPIKey = r.Header.Get("x-api-key")

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	var out struct {
		OK bool `json:"ok"`
	}
	body := map[string]any{"foo": "bar"}
	if err := client.do(context.Background(), http.MethodPost, "/api/test?x=1", body, &out); err != nil {
		t.Fatalf("do error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/api/test?x=1" {
		t.Errorf("path = %q, want %q", gotPath, "/api/test?x=1")
	}
	if gotAPIKey != "test-key" {
		t.Errorf("x-api-key = %q, want %q", gotAPIKey, "test-key")
	}
	if gotBody["foo"] != "bar" {
		t.Errorf("body.foo = %v, want %v", gotBody["foo"], "bar")
	}
	if !out.OK {
		t.Errorf("out.OK = false, want true")
	}
}

func TestClientDo_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "key")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	err = client.do(context.Background(), http.MethodGet, "/fail", nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, err) { // just ensure it's non-nil; concrete type not important
		return
	}
}
