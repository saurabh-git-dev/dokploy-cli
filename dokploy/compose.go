package dokploy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type composeCreateUpdateResponse struct {
	ID        string `json:"id"`
	ComposeID string `json:"composeId"`
}

// GetCompose retrieves a compose app using the official Dokploy
// GET /api/compose.one?composeId=... endpoint. Lookup by name is not
// supported by that endpoint, so an id is required.
func GetCompose(ctx context.Context, client *Client, id, name string) (map[string]any, error) {
	if id == "" {
		return nil, fmt.Errorf("Dokploy compose.one API requires --id; lookup by --name is not supported")
	}
	q := url.Values{}
	q.Set("composeId", id)
	path := "/api/compose.one?" + q.Encode()
	var out map[string]any
	if err := client.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateOrUpdateCompose maps to Dokploy's compose.create and compose.update APIs.
// If id is empty, it calls POST /api/compose.create; otherwise it calls
// POST /api/compose.update with composeId.
func CreateOrUpdateCompose(ctx context.Context, client *Client, id, name, environmentID, composeContent string, envVars map[string]string) (string, error) {
	// Dokploy expects env as a single string; join KEY=VALUE pairs.
	var envLines []string
	for k, v := range envVars {
		envLines = append(envLines, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(envLines)
	envString := strings.Join(envLines, "\n")

	if id == "" {
		payload := map[string]any{
			"name":          name,
			"environmentId": environmentID,
		}
		if composeContent != "" {
			payload["composeFile"] = composeContent
		}
		if envString != "" {
			payload["env"] = envString
		}

		var resp composeCreateUpdateResponse
		if err := client.do(ctx, http.MethodPost, "/api/compose.create", payload, &resp); err != nil {
			return "", err
		}
		if resp.ID != "" {
			return resp.ID, nil
		}
		if resp.ComposeID != "" {
			return resp.ComposeID, nil
		}
		return "", nil
	}

	payload := map[string]any{
		"composeId": id,
	}
	if name != "" {
		payload["name"] = name
	}
	if composeContent != "" {
		payload["composeFile"] = composeContent
	}
	if envString != "" {
		payload["env"] = envString
	}

	var resp composeCreateUpdateResponse
	if err := client.do(ctx, http.MethodPost, "/api/compose.update", payload, &resp); err != nil {
		return "", err
	}
	if resp.ID != "" {
		return resp.ID, nil
	}
	if resp.ComposeID != "" {
		return resp.ComposeID, nil
	}
	return id, nil
}

// DeleteCompose calls POST /api/compose.delete with configurable deleteVolumes.
func DeleteCompose(ctx context.Context, client *Client, id string, deleteVolumes bool) error {
	payload := map[string]any{
		"composeId":     id,
		"deleteVolumes": deleteVolumes,
	}
	return client.do(ctx, http.MethodPost, "/api/compose.delete", payload, nil)
}

// DeployCompose calls POST /api/compose.deploy with the composeId.
func DeployCompose(ctx context.Context, client *Client, id string) error {
	payload := map[string]any{
		"composeId": id,
	}
	return client.do(ctx, http.MethodPost, "/api/compose.deploy", payload, nil)
}
