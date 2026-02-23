package cmd

import (
	"fmt"
	"os"

	"github.com/rainbend/appstoreconnect-cli/internal/api"
	"github.com/rainbend/appstoreconnect-cli/internal/metadata"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local metadata from App Store Connect",
	Long: `Download promotional text, description, keywords and what's new
from the latest version on App Store Connect, organized by platform and locale.

The metadata is saved to:
  .metadata/{platform}/{locale}/description.txt
  .metadata/{platform}/{locale}/keywords.txt
  .metadata/{platform}/{locale}/promotional_text.txt
  .metadata/{platform}/{locale}/whats_new.txt`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringP("app-id", "a", os.Getenv("ASC_APP_ID"), "App Store Connect App ID (env: ASC_APP_ID)")
	initCmd.Flags().StringP("platform", "p", "ios", "Platform: ios or macos")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
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

	client := api.NewClient(keyID, issuerID, privateKeyPath)

	fmt.Printf("Fetching latest %s version...\n", platformStr)
	versions, err := client.ListAppStoreVersions(appID, platform)
	if err != nil {
		return fmt.Errorf("listing versions: %w", err)
	}
	if len(versions) == 0 {
		return fmt.Errorf("no versions found for platform %s", platformStr)
	}

	version := versions[0]
	fmt.Printf("Found version %s (%s)\n", version.Attributes.VersionString, version.Attributes.AppStoreState)

	localizations, err := client.ListVersionLocalizations(version.ID)
	if err != nil {
		return fmt.Errorf("listing localizations: %w", err)
	}
	if len(localizations) == 0 {
		return fmt.Errorf("no localizations found for version %s", version.Attributes.VersionString)
	}

	fmt.Printf("Saving %d localization(s)...\n", len(localizations))

	for _, loc := range localizations {
		m := metadata.Localization{
			Description:     loc.Attributes.Description,
			Keywords:        loc.Attributes.Keywords,
			PromotionalText: loc.Attributes.PromotionalText,
			WhatsNew:        loc.Attributes.WhatsNew,
		}
		if err := metadata.Write(platformStr, loc.Attributes.Locale, m); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write %s: %v\n", loc.Attributes.Locale, err)
			continue
		}
		fmt.Printf("  %s\n", loc.Attributes.Locale)
	}

	fmt.Println("Done! Metadata saved to .metadata/ directory.")
	return nil
}
