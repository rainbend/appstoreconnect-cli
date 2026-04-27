package buildinfo

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestReadIPAExtractsVersionInfo(t *testing.T) {
	tmpDir := t.TempDir()
	ipaPath := filepath.Join(tmpDir, "MyApp.ipa")

	file, err := os.Create(ipaPath)
	if err != nil {
		t.Fatalf("create ipa: %v", err)
	}

	zipWriter := zip.NewWriter(file)
	writer, err := zipWriter.Create("Payload/MyApp.app/Info.plist")
	if err != nil {
		t.Fatalf("create info plist entry: %v", err)
	}
	_, err = writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "https://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleIdentifier</key>
  <string>com.example.myapp</string>
  <key>CFBundleShortVersionString</key>
  <string>1.2.3</string>
  <key>CFBundleVersion</key>
  <string>45</string>
</dict>
</plist>`))
	if err != nil {
		t.Fatalf("write info plist: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close ipa: %v", err)
	}

	info, err := ReadIPA(ipaPath)
	if err != nil {
		t.Fatalf("ReadIPA returned error: %v", err)
	}

	if info.BundleID != "com.example.myapp" {
		t.Fatalf("BundleID = %q, want com.example.myapp", info.BundleID)
	}
	if info.VersionString != "1.2.3" {
		t.Fatalf("VersionString = %q, want 1.2.3", info.VersionString)
	}
	if info.BuildNumber != "45" {
		t.Fatalf("BuildNumber = %q, want 45", info.BuildNumber)
	}
}
