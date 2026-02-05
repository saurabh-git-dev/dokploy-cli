package dokploy

import (
	"context"
	"net/http"
)

// Environment create: POST /api/environment.create
// Environment remove: POST /api/environment.remove

type environmentCreateResponse struct {
	EnvironmentID string `json:"environmentId"`
}

func CreateEnvironment(ctx context.Context, client *Client, _ string, name, projectID string) (string, error) {
	payload := map[string]any{
		"name":      name,
		"projectId": projectID,
	}
	var resp environmentCreateResponse
	if err := client.do(ctx, http.MethodPost, "/api/environment.create", payload, &resp); err != nil {
		return "", err
	}
	// If the API does not return an ID, this will be empty.
	return resp.EnvironmentID, nil
}

func DeleteEnvironment(ctx context.Context, client *Client, id string) error {
	payload := map[string]any{
		"environmentId": id,
	}
	return client.do(ctx, http.MethodPost, "/api/environment.remove", payload, nil)
}
