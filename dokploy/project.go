package dokploy

import (
	"context"
	"errors"
	"net/http"
)

// Project create: POST /api/project.create
// Project remove: POST /api/project.remove

// projectCreateResponse reflects the real Dokploy response, which returns
// both the created project and its default environment.
type projectCreateResponse struct {
	Project struct {
		ProjectID      string  `json:"projectId"`
		Name           string  `json:"name"`
		Description    *string `json:"description"`
		CreatedAt      string  `json:"createdAt"`
		OrganizationID string  `json:"organizationId"`
		Env            string  `json:"env"`
	} `json:"project"`
	Environment struct {
		EnvironmentID string  `json:"environmentId"`
		Name          string  `json:"name"`
		Description   *string `json:"description"`
		CreatedAt     string  `json:"createdAt"`
		Env           string  `json:"env"`
		ProjectID     string  `json:"projectId"`
	} `json:"environment"`
}

// CreateProject calls Dokploy project.create and returns both the
// created project ID and the ID of the default environment Dokploy creates
// alongside it.
func CreateProject(ctx context.Context, client *Client, name string, description string, environment string) (projectID, environmentID string, err error) {
	payload := map[string]any{
		"name":        name,
		"description": description,
		"env":         environment,
	}
	var resp projectCreateResponse
	if err := client.do(ctx, http.MethodPost, "/api/project.create", payload, &resp); err != nil {
		return "", "", err
	}
	return resp.Project.ProjectID, resp.Environment.EnvironmentID, nil
}

func DeleteProject(ctx context.Context, client *Client, id string) error {
	payload := map[string]any{
		"projectId": id,
	}
	return client.do(ctx, http.MethodPost, "/api/project.remove", payload, nil)
}

// Project represents a Dokploy project as returned by project.all.
type Project struct {
	ProjectID      string               `json:"projectId"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	CreatedAt      string               `json:"createdAt"`
	OrganizationID string               `json:"organizationId"`
	Env            string               `json:"env"`
	Environments   []ProjectEnvironment `json:"environments"`
}

// ProjectEnvironment represents an environment embedded in a project
// response from project.all.
type ProjectEnvironment struct {
	EnvironmentID string `json:"environmentId"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	CreatedAt     string `json:"createdAt"`
}

// ListProjects calls GET /api/project.all and returns all projects.
func ListProjects(ctx context.Context, client *Client) ([]Project, error) {
	var out []Project
	if err := client.do(ctx, http.MethodGet, "/api/project.all", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProject finds a project by name and one of its environments by envName,
// using the data returned from ListProjects. It returns the matching
// projectId and environmentId.
func GetProject(ctx context.Context, client *Client, name, envName string) (string, string, error) {
	if name == "" {
		return "", "", errors.New("project name is required")
	}
	if envName == "" {
		return "", "", errors.New("environment name is required")
	}

	projects, err := ListProjects(ctx, client)
	if err != nil {
		return "", "", err
	}

	for _, p := range projects {
		if p.Name != name {
			continue
		}
		for _, e := range p.Environments {
			if e.Name == envName {
				return p.ProjectID, e.EnvironmentID, nil
			}
		}
		return "", "", errors.New("environment not found for project")
	}

	return "", "", errors.New("project not found")
}
