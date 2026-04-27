package buildinfo

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AppInfo struct {
	BundleID      string
	VersionString string
	BuildNumber   string
}

func ReadIPA(path string) (*AppInfo, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("opening ipa: %w", err)
	}
	defer reader.Close()

	var infoPlist *zip.File
	for _, file := range reader.File {
		if isAppInfoPlist(file.Name) {
			infoPlist = file
			break
		}
	}
	if infoPlist == nil {
		return nil, fmt.Errorf("Info.plist not found in Payload/*.app")
	}

	rc, err := infoPlist.Open()
	if err != nil {
		return nil, fmt.Errorf("opening Info.plist: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("reading Info.plist: %w", err)
	}

	info, err := parseXMLInfoPlist(data)
	if err == nil && info.VersionString != "" && info.BuildNumber != "" {
		return info, nil
	}

	info, err = parseInfoPlistWithPlutil(data)
	if err != nil {
		return nil, fmt.Errorf("parsing Info.plist: %w", err)
	}
	if info.VersionString == "" {
		return nil, fmt.Errorf("CFBundleShortVersionString not found in Info.plist")
	}
	if info.BuildNumber == "" {
		return nil, fmt.Errorf("CFBundleVersion not found in Info.plist")
	}

	return info, nil
}

func isAppInfoPlist(name string) bool {
	if !strings.HasPrefix(name, "Payload/") || !strings.HasSuffix(name, ".app/Info.plist") {
		return false
	}
	parts := strings.Split(name, "/")
	return len(parts) == 3
}

func parseXMLInfoPlist(data []byte) (*AppInfo, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	values := make(map[string]string)
	var lastKey string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		switch start.Name.Local {
		case "key":
			var value string
			if err := decoder.DecodeElement(&value, &start); err != nil {
				return nil, err
			}
			lastKey = value
		case "string", "integer":
			if lastKey == "" {
				continue
			}
			var value string
			if err := decoder.DecodeElement(&value, &start); err != nil {
				return nil, err
			}
			values[lastKey] = value
			lastKey = ""
		}
	}

	return appInfoFromMap(values), nil
}

func parseInfoPlistWithPlutil(data []byte) (*AppInfo, error) {
	dir, err := os.MkdirTemp("", "asctl-info-plist-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	plistPath := filepath.Join(dir, "Info.plist")
	if err := os.WriteFile(plistPath, data, 0600); err != nil {
		return nil, fmt.Errorf("writing temp Info.plist: %w", err)
	}

	output, err := exec.Command("plutil", "-convert", "json", "-o", "-", plistPath).Output()
	if err != nil {
		return nil, fmt.Errorf("converting plist with plutil: %w", err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(output, &values); err != nil {
		return nil, fmt.Errorf("decoding plutil JSON: %w", err)
	}

	return appInfoFromAnyMap(values), nil
}

func appInfoFromMap(values map[string]string) *AppInfo {
	return &AppInfo{
		BundleID:      values["CFBundleIdentifier"],
		VersionString: values["CFBundleShortVersionString"],
		BuildNumber:   values["CFBundleVersion"],
	}
}

func appInfoFromAnyMap(values map[string]interface{}) *AppInfo {
	stringValue := func(key string) string {
		switch value := values[key].(type) {
		case string:
			return value
		case float64:
			return fmt.Sprintf("%.0f", value)
		default:
			return ""
		}
	}

	return &AppInfo{
		BundleID:      stringValue("CFBundleIdentifier"),
		VersionString: stringValue("CFBundleShortVersionString"),
		BuildNumber:   stringValue("CFBundleVersion"),
	}
}
