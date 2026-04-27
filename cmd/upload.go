package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rainbend/appstoreconnect-cli/internal/api"
	"github.com/rainbend/appstoreconnect-cli/internal/buildinfo"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a build to App Store Connect",
	Long: `Upload an app build using the App Store Connect Build Upload API.

The command creates a build upload, reserves the build file, uploads all
parts returned by App Store Connect, and commits the upload for processing.

For .ipa files, version and build number are read from Payload/*.app/Info.plist
when --version or --build are omitted. For .pkg files, pass both values.`,
	RunE: runUpload,
}

func init() {
	uploadCmd.Flags().StringP("app-id", "a", os.Getenv("ASC_APP_ID"), "App Store Connect App ID (env: ASC_APP_ID)")
	uploadCmd.Flags().StringP("platform", "p", "ios", "Platform: ios or macos")
	uploadCmd.Flags().StringP("file", "f", "", "Build file to upload (.ipa or .pkg)")
	uploadCmd.Flags().StringP("version", "v", "", "CFBundleShortVersionString, e.g. 1.2.0 (auto-read for .ipa)")
	uploadCmd.Flags().StringP("build", "b", "", "CFBundleVersion/build number, e.g. 42 (auto-read for .ipa)")
	uploadCmd.Flags().String("uti", "", "Build file UTI override (defaults by extension)")
	uploadCmd.Flags().Bool("wait", false, "Wait for App Store Connect to finish processing the build upload")
	uploadCmd.Flags().Duration("wait-timeout", 30*time.Minute, "Maximum time to wait when --wait is set")
	_ = uploadCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	if err := validateAuthFlags(); err != nil {
		return err
	}

	appID, _ := cmd.Flags().GetString("app-id")
	if appID == "" {
		return fmt.Errorf("--app-id or ASC_APP_ID environment variable is required")
	}

	platformStr, _ := cmd.Flags().GetString("platform")
	platform, err := api.ParsePlatform(platformStr)
	if err != nil {
		return err
	}

	filePath, _ := cmd.Flags().GetString("file")
	if filePath == "" {
		return fmt.Errorf("--file is required")
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat build file: %w", err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("build file cannot be a directory: %s", filePath)
	}

	versionString, _ := cmd.Flags().GetString("version")
	buildNumber, _ := cmd.Flags().GetString("build")
	if versionString == "" || buildNumber == "" {
		if strings.EqualFold(filepath.Ext(filePath), ".ipa") {
			info, err := buildinfo.ReadIPA(filePath)
			if err != nil {
				return fmt.Errorf("reading version from ipa: %w; pass --version and --build to set them explicitly", err)
			}
			if versionString == "" {
				versionString = info.VersionString
			}
			if buildNumber == "" {
				buildNumber = info.BuildNumber
			}
			if info.BundleID != "" {
				fmt.Printf("Found bundle %s, version %s (%s)\n", info.BundleID, versionString, buildNumber)
			}
		}
	}
	if versionString == "" {
		return fmt.Errorf("--version is required when it cannot be read from the build file")
	}
	if buildNumber == "" {
		return fmt.Errorf("--build is required when it cannot be read from the build file")
	}

	uti, _ := cmd.Flags().GetString("uti")
	if uti == "" {
		uti, err = inferBuildFileUTI(filePath)
		if err != nil {
			return err
		}
	}

	waitForProcessing, _ := cmd.Flags().GetBool("wait")
	waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
	if waitForProcessing && waitTimeout <= 0 {
		return fmt.Errorf("--wait-timeout must be greater than 0")
	}

	client := api.NewClient(keyID, issuerID, privateKeyPath)

	fmt.Printf("Creating build upload for %s (%s) on %s...\n", versionString, buildNumber, platformStr)
	buildUpload, err := client.CreateBuildUpload(appID, versionString, buildNumber, platform)
	if err != nil {
		return fmt.Errorf("creating build upload: %w", err)
	}
	fmt.Printf("Build Upload ID: %s\n", buildUpload.ID)

	fileName := filepath.Base(filePath)
	fmt.Printf("Reserving build file %s (%s)...\n", fileName, formatBytes(fileInfo.Size()))
	buildUploadFile, err := client.CreateBuildUploadFile(buildUpload.ID, fileName, fileInfo.Size(), uti)
	if err != nil {
		return fmt.Errorf("creating build upload file: %w", err)
	}

	operations := buildUploadFile.Attributes.UploadOperations
	fmt.Printf("Uploading %d part(s)...\n", len(operations))
	err = client.UploadBuildFile(filePath, operations, func(index, total int, op api.UploadOperation) {
		fmt.Printf("  Part %d/%d: offset %d, %s\n", index, total, op.Offset, formatBytes(op.Length))
	})
	if err != nil {
		return err
	}

	fmt.Println("Committing upload...")
	committedFile, err := client.CommitBuildUploadFile(buildUploadFile.ID)
	if err != nil {
		return fmt.Errorf("committing build upload file: %w", err)
	}
	printAssetDeliveryMessages(committedFile.Attributes.AssetDeliveryState)

	if waitForProcessing {
		if err := waitForBuildUpload(client, buildUpload.ID, waitTimeout); err != nil {
			return err
		}
		fmt.Println("Build upload processing complete!")
	} else {
		fmt.Println("Upload complete! App Store Connect will continue processing the build.")
	}

	return nil
}

func inferBuildFileUTI(filePath string) (string, error) {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".ipa":
		return "com.apple.ipa", nil
	case ".pkg":
		return "com.apple.pkg", nil
	default:
		return "", fmt.Errorf("cannot infer UTI for %s; pass --uti explicitly", filePath)
	}
}

func waitForBuildUpload(client *api.Client, buildUploadID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		buildUpload, err := client.GetBuildUpload(buildUploadID)
		if err != nil {
			return fmt.Errorf("checking build upload status: %w", err)
		}

		state := strings.ToUpper(buildUpload.Attributes.State)
		if state == "" {
			state = "UNKNOWN"
		}
		fmt.Printf("  Build upload state: %s\n", state)

		switch state {
		case "COMPLETE":
			return nil
		case "FAILED":
			return fmt.Errorf("build upload failed")
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for build upload %s after %s", buildUploadID, timeout)
		}
		time.Sleep(30 * time.Second)
	}
}

func printAssetDeliveryMessages(state *api.AssetDeliveryState) {
	if state == nil {
		return
	}
	if state.State != "" {
		fmt.Printf("Asset delivery state: %s\n", state.State)
	}
	for _, warning := range state.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", deliveryMessageText(warning))
	}
	for _, errMsg := range state.Errors {
		fmt.Fprintf(os.Stderr, "Error: %s\n", deliveryMessageText(errMsg))
	}
}

func deliveryMessageText(message api.AssetDeliveryMessage) string {
	for _, value := range []string{message.Detail, message.Message, message.Title, message.Code} {
		if value != "" {
			return value
		}
	}
	return "unknown delivery message"
}

func formatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	value := float64(size)
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}

	return fmt.Sprintf("%.1f %s", value, units[unit])
}
