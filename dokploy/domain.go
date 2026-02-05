package dokploy

import (
	"context"
	"net/http"
	"net/url"
)

// Domain create: POST /api/domain.create
// Domain update: POST /api/domain.update
// Domain delete: POST /api/domain.delete
// Domain by compose: GET /api/domain.byComposeId?composeId=...

type domainCreateUpdateResponse struct {
	DomainID string `json:"domainId"`
}

type domainByComposeItem struct {
	DomainID string `json:"domainId"`
	Host     string `json:"host"`
	Path     string `json:"path"`
}

// findExistingDomainIDByCompose lists domains for a compose and returns
// an existing domainId if any. To avoid duplicates, we prefer a domain
// matching the same host and path; if none match but some exist, we
// return the first.
func findExistingDomainIDByCompose(ctx context.Context, client *Client, composeID, host, path string) (string, error) {
	if composeID == "" {
		return "", nil
	}
	q := url.Values{}
	q.Set("composeId", composeID)
	endpoint := "/api/domain.byComposeId?" + q.Encode()
	var items []domainByComposeItem
	if err := client.do(ctx, http.MethodGet, endpoint, nil, &items); err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", nil
	}
	for _, it := range items {
		if it.Host == host && it.Path == path {
			return it.DomainID, nil
		}
	}
	return items[0].DomainID, nil
}

func CreateOrUpdateDomain(
	ctx context.Context,
	client *Client,
	id string,
	host string,
	path string,
	port int,
	serviceName string,
	composeID string,
	certificateType string,
	https bool,
) (string, error) {
	payload := map[string]any{
		"host":            host,
		"path":            path,
		"port":            port,
		"serviceName":     serviceName,
		"composeId":       composeID,
		"certificateType": certificateType,
		"https":           https,
		"domainType":      "compose",
	}

	// If no explicit id is provided, try to find an existing domain for
	// this compose (and host/path) to update instead of creating a duplicate.
	if id == "" {
		var err error
		id, err = findExistingDomainIDByCompose(ctx, client, composeID, host, path)
		if err != nil {
			return "", err
		}
	}

	var resp domainCreateUpdateResponse
	if id != "" {
		payload["domainId"] = id
		if err := client.do(ctx, http.MethodPost, "/api/domain.update", payload, &resp); err != nil {
			return "", err
		}
	} else {
		if err := client.do(ctx, http.MethodPost, "/api/domain.create", payload, &resp); err != nil {
			return "", err
		}
	}
	return resp.DomainID, nil
}

func DeleteDomain(ctx context.Context, client *Client, id string) error {
	payload := map[string]any{
		"domainId": id,
	}
	return client.do(ctx, http.MethodPost, "/api/domain.delete", payload, nil)
}
