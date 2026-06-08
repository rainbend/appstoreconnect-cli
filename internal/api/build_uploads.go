package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const buildUploadAssetType = "ASSET"

func (c *Client) CreateBuildUpload(appID, versionString, buildNumber string, platform Platform) (*BuildUpload, error) {
	req := CreateBuildUploadRequest{
		Data: CreateBuildUploadData{
			Type: "buildUploads",
			Attributes: BuildUploadAttributes{
				CFBundleShortVersionString: versionString,
				CFBundleVersion:            buildNumber,
				Platform:                   string(platform),
			},
			Relationships: CreateBuildUploadRelationships{
				App: RelationshipData{
					Data: ResourceIdentifier{
						Type: "apps",
						ID:   appID,
					},
				},
			},
		},
	}

	data, err := c.do("POST", "/buildUploads", req)
	if err != nil {
		return nil, err
	}

	var resp SingleResponse[BuildUpload]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &resp.Data, nil
}

func (c *Client) GetBuildUpload(buildUploadID string) (*BuildUpload, error) {
	data, err := c.do("GET", fmt.Sprintf("/buildUploads/%s", buildUploadID), nil)
	if err != nil {
		return nil, err
	}

	var resp SingleResponse[BuildUpload]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &resp.Data, nil
}

func (c *Client) CreateBuildUploadFile(buildUploadID, fileName string, fileSize int64, uti string) (*BuildUploadFile, error) {
	req := CreateBuildUploadFileRequest{
		Data: CreateBuildUploadFileData{
			Type: "buildUploadFiles",
			Attributes: BuildUploadFileAttributes{
				FileName:  fileName,
				FileSize:  fileSize,
				AssetType: buildUploadAssetType,
				UTI:       uti,
			},
			Relationships: CreateBuildUploadFileRelationships{
				BuildUpload: RelationshipData{
					Data: ResourceIdentifier{
						Type: "buildUploads",
						ID:   buildUploadID,
					},
				},
			},
		},
	}

	data, err := c.do("POST", "/buildUploadFiles", req)
	if err != nil {
		return nil, err
	}

	var resp SingleResponse[BuildUploadFile]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &resp.Data, nil
}

func (c *Client) CommitBuildUploadFile(buildUploadFileID string) (*BuildUploadFile, error) {
	req := UpdateBuildUploadFileRequest{
		Data: UpdateBuildUploadFileData{
			Type: "buildUploadFiles",
			ID:   buildUploadFileID,
			Attributes: BuildUploadFileAttributes{
				Uploaded: true,
			},
		},
	}

	data, err := c.do("PATCH", fmt.Sprintf("/buildUploadFiles/%s", buildUploadFileID), req)
	if err != nil {
		return nil, err
	}

	var resp SingleResponse[BuildUploadFile]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &resp.Data, nil
}

func (c *Client) UploadBuildFile(filePath string, operations []UploadOperation, progress func(index, total int, op UploadOperation)) error {
	if len(operations) == 0 {
		return fmt.Errorf("no upload operations returned by App Store Connect")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening build file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat build file: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	total := len(operations)
	for i, op := range operations {
		if op.Length <= 0 {
			return fmt.Errorf("upload operation %d has invalid length %d", i+1, op.Length)
		}
		if op.Offset < 0 || op.Offset+op.Length > info.Size() {
			return fmt.Errorf("upload operation %d range is outside file bounds", i+1)
		}
		if op.URL == "" {
			return fmt.Errorf("upload operation %d is missing URL", i+1)
		}

		if progress != nil {
			progress(i+1, total, op)
		}
		if err := uploadBuildFilePart(client, file, op); err != nil {
			return fmt.Errorf("uploading part %d/%d: %w", i+1, total, err)
		}
	}

	return nil
}

func uploadBuildFilePart(client *http.Client, file *os.File, op UploadOperation) error {
	method := op.Method
	if method == "" {
		method = http.MethodPut
	}

	body := io.NewSectionReader(file, op.Offset, op.Length)
	req, err := http.NewRequest(method, op.URL, body)
	if err != nil {
		return fmt.Errorf("creating upload request: %w", err)
	}
	req.ContentLength = op.Length

	for _, header := range op.RequestHeaders {
		if header.Name == "" {
			continue
		}
		req.Header.Set(header.Name, header.Value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("upload error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
