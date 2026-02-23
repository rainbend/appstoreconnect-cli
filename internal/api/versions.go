package api

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ListAppStoreVersions(appID string, platform Platform) ([]AppStoreVersion, error) {
	path := fmt.Sprintf("/apps/%s/appStoreVersions?filter[platform]=%s", appID, platform)

	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListResponse[AppStoreVersion]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return resp.Data, nil
}

func (c *Client) CreateAppStoreVersion(appID, versionString string, platform Platform) (*AppStoreVersion, error) {
	req := CreateVersionRequest{
		Data: CreateVersionData{
			Type: "appStoreVersions",
			Attributes: AppStoreVersionAttributes{
				VersionString: versionString,
				Platform:      string(platform),
			},
			Relationships: CreateVersionRelationships{
				App: RelationshipData{
					Data: ResourceIdentifier{
						Type: "apps",
						ID:   appID,
					},
				},
			},
		},
	}

	data, err := c.do("POST", "/appStoreVersions", req)
	if err != nil {
		return nil, err
	}

	var resp SingleResponse[AppStoreVersion]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &resp.Data, nil
}
