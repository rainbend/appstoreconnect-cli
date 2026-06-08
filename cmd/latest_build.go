package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rainbend/appstoreconnect-cli/internal/api"
	"github.com/spf13/cobra"
)

var latestBuildCmd = &cobra.Command{
	Use:   "latest-build",
	Short: "Print the latest build number for a platform",
	Long: `Print the latest App Store Connect build number for an app platform.

By default the command prints only Build.Attributes.version so it can be used
from scripts. Use --details to print the build ID, processing state, and upload
date as well.`,
	RunE: runLatestBuild,
}

func init() {
	latestBuildCmd.Flags().StringP("app-id", "a", os.Getenv("ASC_APP_ID"), "App Store Connect App ID (env: ASC_APP_ID)")
	latestBuildCmd.Flags().StringP("platform", "p", "ios", "Platform: ios or macos")
	latestBuildCmd.Flags().String("state", "all", "Processing state: all, valid, processing, failed, or invalid")
	latestBuildCmd.Flags().Bool("details", false, "Print build ID, state, and upload date")
	rootCmd.AddCommand(latestBuildCmd)
}

func runLatestBuild(cmd *cobra.Command, args []string) error {
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

	stateFlag, _ := cmd.Flags().GetString("state")
	state, err := normalizeBuildProcessingState(stateFlag)
	if err != nil {
		return err
	}

	client := api.NewClient(keyID, issuerID, privateKeyPath)
	build, err := client.LatestBuild(appID, platform, state)
	if err != nil {
		return fmt.Errorf("fetching latest build: %w", err)
	}
	if build == nil {
		stateText := "any processing state"
		if state != "" {
			stateText = state
		}
		return fmt.Errorf("no builds found for platform %s with %s", platformStr, stateText)
	}

	details, _ := cmd.Flags().GetBool("details")
	if !details {
		fmt.Println(build.Attributes.Version)
		return nil
	}

	fmt.Printf("Build: %s\n", build.Attributes.Version)
	fmt.Printf("ID: %s\n", build.ID)
	if build.Attributes.ProcessingState != "" {
		fmt.Printf("State: %s\n", build.Attributes.ProcessingState)
	}
	if build.Attributes.UploadedDate != "" {
		fmt.Printf("Uploaded: %s\n", build.Attributes.UploadedDate)
	}
	if build.Attributes.MinOSVersion != "" {
		fmt.Printf("Min OS: %s\n", build.Attributes.MinOSVersion)
	}
	if build.Attributes.ExpirationDate != "" {
		fmt.Printf("Expires: %s\n", build.Attributes.ExpirationDate)
	}
	if build.Attributes.Expired {
		fmt.Println("Expired: true")
	}

	return nil
}

func normalizeBuildProcessingState(state string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "", "all":
		return "", nil
	case "valid":
		return "VALID", nil
	case "processing":
		return "PROCESSING", nil
	case "failed":
		return "FAILED", nil
	case "invalid":
		return "INVALID", nil
	default:
		return "", fmt.Errorf("invalid --state %q: must be all, valid, processing, failed, or invalid", state)
	}
}
