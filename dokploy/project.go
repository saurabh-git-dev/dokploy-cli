package dokploy

import (
	"context"
	"net/http"
)

// Project create: POST /api/project.create
// Project remove: POST /api/project.remove

type projectCreateResponse struct {
	ProjectID string `json:"projectId"`
}

func CreateProject(ctx context.Context, client *Client, _ string, name string) (string, error) {
	payload := map[string]any{
		"name": name,
	}
	var resp projectCreateResponse
	if err := client.do(ctx, http.MethodPost, "/api/project.create", payload, &resp); err != nil {
		return "", err
	}
	// If the API does not return an ID, this will be empty.
	return resp.ProjectID, nil
}

func DeleteProject(ctx context.Context, client *Client, id string) error {
	payload := map[string]any{
		"projectId": id,
	}
	return client.do(ctx, http.MethodPost, "/api/project.remove", payload, nil)
}
