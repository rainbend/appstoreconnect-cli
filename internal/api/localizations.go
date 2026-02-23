package api

import (
	"encoding/json"
	"fmt"
)

func (c *Client) ListVersionLocalizations(versionID string) ([]AppStoreVersionLocalization, error) {
	path := fmt.Sprintf("/appStoreVersions/%s/appStoreVersionLocalizations", versionID)

	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListResponse[AppStoreVersionLocalization]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return resp.Data, nil
}

func (c *Client) UpdateVersionLocalization(localizationID string, attrs AppStoreVersionLocalizationAttributes) error {
	req := UpdateLocalizationRequest{
		Data: UpdateLocalizationData{
			Type:       "appStoreVersionLocalizations",
			ID:         localizationID,
			Attributes: attrs,
		},
	}

	_, err := c.do("PATCH", fmt.Sprintf("/appStoreVersionLocalizations/%s", localizationID), req)
	return err
}

func (c *Client) CreateVersionLocalization(versionID, locale string, attrs AppStoreVersionLocalizationAttributes) error {
	attrs.Locale = locale
	req := CreateLocalizationRequest{
		Data: CreateLocalizationData{
			Type:       "appStoreVersionLocalizations",
			Attributes: attrs,
			Relationships: CreateLocalizationRelationships{
				AppStoreVersion: RelationshipData{
					Data: ResourceIdentifier{
						Type: "appStoreVersions",
						ID:   versionID,
					},
				},
			},
		},
	}

	_, err := c.do("POST", "/appStoreVersionLocalizations", req)
	return err
}
