package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) LatestBuild(appID string, platform Platform, processingState string) (*Build, error) {
	values := url.Values{}
	values.Set("filter[app]", appID)
	values.Set("filter[preReleaseVersion.platform]", string(platform))
	values.Set("fields[builds]", "version,uploadedDate,expirationDate,expired,minOsVersion,processingState,buildAudienceType,usesNonExemptEncryption")
	values.Set("sort", "-uploadedDate")
	values.Set("limit", "1")
	if processingState != "" {
		values.Set("filter[processingState]", strings.ToUpper(processingState))
	}

	data, err := c.do("GET", "/builds?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}

	var resp ListResponse[Build]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}

	return &resp.Data[0], nil
}
