package api

import "fmt"

type Platform string

const (
	PlatformIOS   Platform = "IOS"
	PlatformMacOS Platform = "MAC_OS"
)

func ParsePlatform(s string) (Platform, error) {
	switch s {
	case "ios":
		return PlatformIOS, nil
	case "macos":
		return PlatformMacOS, nil
	default:
		return "", fmt.Errorf("invalid platform %q: must be 'ios' or 'macos'", s)
	}
}

type ListResponse[T any] struct {
	Data  []T   `json:"data"`
	Links Links `json:"links,omitempty"`
}

type SingleResponse[T any] struct {
	Data T `json:"data"`
}

type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
}

type AppStoreVersion struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes AppStoreVersionAttributes `json:"attributes"`
}

type AppStoreVersionAttributes struct {
	VersionString string `json:"versionString"`
	Platform      string `json:"platform"`
	AppStoreState string `json:"appStoreState,omitempty"`
}

type AppStoreVersionLocalization struct {
	Type       string                                `json:"type"`
	ID         string                                `json:"id"`
	Attributes AppStoreVersionLocalizationAttributes `json:"attributes"`
}

type AppStoreVersionLocalizationAttributes struct {
	Locale          string `json:"locale,omitempty"`
	Description     string `json:"description,omitempty"`
	Keywords        string `json:"keywords,omitempty"`
	PromotionalText string `json:"promotionalText,omitempty"`
	WhatsNew        string `json:"whatsNew,omitempty"`
}

type CreateVersionRequest struct {
	Data CreateVersionData `json:"data"`
}

type CreateVersionData struct {
	Type          string                     `json:"type"`
	Attributes    AppStoreVersionAttributes  `json:"attributes"`
	Relationships CreateVersionRelationships `json:"relationships"`
}

type CreateVersionRelationships struct {
	App RelationshipData `json:"app"`
}

type RelationshipData struct {
	Data ResourceIdentifier `json:"data"`
}

type ResourceIdentifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CreateLocalizationRequest struct {
	Data CreateLocalizationData `json:"data"`
}

type CreateLocalizationData struct {
	Type          string                                `json:"type"`
	Attributes    AppStoreVersionLocalizationAttributes `json:"attributes"`
	Relationships CreateLocalizationRelationships       `json:"relationships"`
}

type CreateLocalizationRelationships struct {
	AppStoreVersion RelationshipData `json:"appStoreVersion"`
}

type UpdateLocalizationRequest struct {
	Data UpdateLocalizationData `json:"data"`
}

type UpdateLocalizationData struct {
	Type       string                                `json:"type"`
	ID         string                                `json:"id"`
	Attributes AppStoreVersionLocalizationAttributes `json:"attributes"`
}
