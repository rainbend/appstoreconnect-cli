package api

import (
	"encoding/json"
	"fmt"
)

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

type Build struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Attributes BuildAttributes `json:"attributes"`
}

type BuildAttributes struct {
	Version                 string `json:"version"`
	UploadedDate            string `json:"uploadedDate,omitempty"`
	ExpirationDate          string `json:"expirationDate,omitempty"`
	Expired                 bool   `json:"expired,omitempty"`
	MinOSVersion            string `json:"minOsVersion,omitempty"`
	ProcessingState         string `json:"processingState,omitempty"`
	BuildAudienceType       string `json:"buildAudienceType,omitempty"`
	UsesNonExemptEncryption bool   `json:"usesNonExemptEncryption,omitempty"`
}

type BuildUpload struct {
	Type       string                `json:"type"`
	ID         string                `json:"id"`
	Attributes BuildUploadAttributes `json:"attributes"`
}

type BuildUploadAttributes struct {
	CFBundleShortVersionString string           `json:"cfBundleShortVersionString,omitempty"`
	CFBundleVersion            string           `json:"cfBundleVersion,omitempty"`
	CreatedDate                string           `json:"createdDate,omitempty"`
	State                      BuildUploadState `json:"state,omitempty"`
	Platform                   string           `json:"platform,omitempty"`
	UploadedDate               string           `json:"uploadedDate,omitempty"`
}

type BuildUploadState string

func (s BuildUploadState) String() string {
	return string(s)
}

func (s *BuildUploadState) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		*s = BuildUploadState(value)
		return nil
	}

	var stateObject struct {
		State string `json:"state"`
		Code  string `json:"code"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &stateObject); err != nil {
		return err
	}

	switch {
	case stateObject.State != "":
		*s = BuildUploadState(stateObject.State)
	case stateObject.Code != "":
		*s = BuildUploadState(stateObject.Code)
	default:
		*s = BuildUploadState(stateObject.Value)
	}
	return nil
}

type BuildUploadFile struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes BuildUploadFileAttributes `json:"attributes"`
}

type BuildUploadFileAttributes struct {
	FileName           string              `json:"fileName,omitempty"`
	FileSize           int64               `json:"fileSize,omitempty"`
	AssetType          string              `json:"assetType,omitempty"`
	UTI                string              `json:"uti,omitempty"`
	Uploaded           bool                `json:"uploaded,omitempty"`
	UploadOperations   []UploadOperation   `json:"uploadOperations,omitempty"`
	AssetDeliveryState *AssetDeliveryState `json:"assetDeliveryState,omitempty"`
}

type UploadOperation struct {
	Method         string                  `json:"method"`
	URL            string                  `json:"url"`
	Offset         int64                   `json:"offset"`
	Length         int64                   `json:"length"`
	RequestHeaders []UploadOperationHeader `json:"requestHeaders,omitempty"`
}

type UploadOperationHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AssetDeliveryState struct {
	State    string                 `json:"state,omitempty"`
	Errors   []AssetDeliveryMessage `json:"errors,omitempty"`
	Warnings []AssetDeliveryMessage `json:"warnings,omitempty"`
}

type AssetDeliveryMessage struct {
	Code    string `json:"code,omitempty"`
	Title   string `json:"title,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Message string `json:"message,omitempty"`
}

type CreateBuildUploadRequest struct {
	Data CreateBuildUploadData `json:"data"`
}

type CreateBuildUploadData struct {
	Type          string                         `json:"type"`
	Attributes    BuildUploadAttributes          `json:"attributes"`
	Relationships CreateBuildUploadRelationships `json:"relationships"`
}

type CreateBuildUploadRelationships struct {
	App RelationshipData `json:"app"`
}

type CreateBuildUploadFileRequest struct {
	Data CreateBuildUploadFileData `json:"data"`
}

type CreateBuildUploadFileData struct {
	Type          string                             `json:"type"`
	Attributes    BuildUploadFileAttributes          `json:"attributes"`
	Relationships CreateBuildUploadFileRelationships `json:"relationships"`
}

type CreateBuildUploadFileRelationships struct {
	BuildUpload RelationshipData `json:"buildUpload"`
}

type UpdateBuildUploadFileRequest struct {
	Data UpdateBuildUploadFileData `json:"data"`
}

type UpdateBuildUploadFileData struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes BuildUploadFileAttributes `json:"attributes"`
}
