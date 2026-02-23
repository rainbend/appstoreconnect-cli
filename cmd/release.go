package cmd

import (
	"fmt"
	"os"

	"github.com/rainbend/appstoreconnect-cli/internal/api"
	"github.com/rainbend/appstoreconnect-cli/internal/metadata"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create or update a version on App Store Connect",
	Long: `Create a new app version (or update an existing one) with metadata
from local files. Reads description, keywords, promotional text, and what's new
from the .metadata/ directory.

If the specified version already exists, its localizations will be updated.
Otherwise a new version is created first.`,
	RunE: runRelease,
}

func init() {
	releaseCmd.Flags().StringP("app-id", "a", os.Getenv("ASC_APP_ID"), "App Store Connect App ID (env: ASC_APP_ID)")
	releaseCmd.Flags().StringP("version", "v", "", "Version string, e.g. 1.2.0 (required)")
	releaseCmd.Flags().StringP("platform", "p", "ios", "Platform: ios or macos")
	releaseCmd.Flags().String("whats-new", "", "What's new text (overrides whats_new.txt for all locales)")
	_ = releaseCmd.MarkFlagRequired("version")
	rootCmd.AddCommand(releaseCmd)
}

func runRelease(cmd *cobra.Command, args []string) error {
	if err := validateAuthFlags(); err != nil {
		return err
	}

	appID, _ := cmd.Flags().GetString("app-id")
	if appID == "" {
		return fmt.Errorf("--app-id or ASC_APP_ID environment variable is required")
	}
	versionStr, _ := cmd.Flags().GetString("version")
	platformStr, _ := cmd.Flags().GetString("platform")
	whatsNew, _ := cmd.Flags().GetString("whats-new")

	platform, err := api.ParsePlatform(platformStr)
	if err != nil {
		return err
	}

	client := api.NewClient(keyID, issuerID, privateKeyPath)

	fmt.Printf("Checking for existing version %s...\n", versionStr)
	versions, err := client.ListAppStoreVersions(appID, platform)
	if err != nil {
		return fmt.Errorf("listing versions: %w", err)
	}

	var versionID string
	for _, v := range versions {
		if v.Attributes.VersionString == versionStr {
			versionID = v.ID
			fmt.Printf("Found existing version %s (%s)\n", versionStr, v.Attributes.AppStoreState)
			break
		}
	}

	if versionID == "" {
		fmt.Printf("Creating version %s...\n", versionStr)
		version, err := client.CreateAppStoreVersion(appID, versionStr, platform)
		if err != nil {
			return fmt.Errorf("creating version: %w", err)
		}
		versionID = version.ID
		fmt.Printf("Created version %s\n", versionStr)
	}

	locales, err := metadata.ListLocales(platformStr)
	if err != nil {
		return fmt.Errorf("listing local metadata: %w", err)
	}
	if len(locales) == 0 {
		return fmt.Errorf("no local metadata found in .metadata/%s/", platformStr)
	}

	existingLocs, err := client.ListVersionLocalizations(versionID)
	if err != nil {
		return fmt.Errorf("listing existing localizations: %w", err)
	}

	existingMap := make(map[string]string)
	for _, loc := range existingLocs {
		existingMap[loc.Attributes.Locale] = loc.ID
	}

	for _, locale := range locales {
		m, err := metadata.Read(platformStr, locale)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read metadata for %s: %v\n", locale, err)
			continue
		}

		if whatsNew != "" {
			m.WhatsNew = whatsNew
		}

		attrs := api.AppStoreVersionLocalizationAttributes{
			Description:     m.Description,
			Keywords:        m.Keywords,
			PromotionalText: m.PromotionalText,
			WhatsNew:        m.WhatsNew,
		}

		if locID, exists := existingMap[locale]; exists {
			fmt.Printf("  Updating %s...\n", locale)
			if err := client.UpdateVersionLocalization(locID, attrs); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", locale, err)
				continue
			}
		} else {
			fmt.Printf("  Creating %s...\n", locale)
			if err := client.CreateVersionLocalization(versionID, locale, attrs); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create %s: %v\n", locale, err)
				continue
			}
		}
		fmt.Printf("  Done: %s\n", locale)
	}

	fmt.Println("Release complete!")
	return nil
}
