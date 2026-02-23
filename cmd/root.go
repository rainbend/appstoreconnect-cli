package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var _ = godotenv.Load()

var (
	keyID          string
	issuerID       string
	privateKeyPath string
)

var rootCmd = &cobra.Command{
	Use:     "asctl",
	Short:   "CLI tool for managing App Store Connect metadata",
	Version: version + " (" + commit + " " + date + ")",
	Long: `A CLI tool to manage App Store Connect app metadata.

Supports initializing local metadata from App Store Connect and
creating/updating versions with promotional text, descriptions,
and keywords across multiple languages for iOS and macOS apps.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&keyID, "key-id", os.Getenv("ASC_KEY_ID"), "API Key ID (env: ASC_KEY_ID)")
	rootCmd.PersistentFlags().StringVar(&issuerID, "issuer-id", os.Getenv("ASC_ISSUER_ID"), "Issuer ID (env: ASC_ISSUER_ID)")
	rootCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", os.Getenv("ASC_PRIVATE_KEY"), "Path to private key .p8 file (env: ASC_PRIVATE_KEY, default: ~/.config/appstoreconnect/AuthKey_{KEY_ID}.p8)")
}

func Execute() error {
	return rootCmd.Execute()
}

func validateAuthFlags() error {
	if keyID == "" {
		return fmt.Errorf("--key-id or ASC_KEY_ID environment variable is required")
	}
	if issuerID == "" {
		return fmt.Errorf("--issuer-id or ASC_ISSUER_ID environment variable is required")
	}
	if privateKeyPath == "" {
		if keyID != "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cannot determine home directory: %w", err)
			}
			privateKeyPath = filepath.Join(home, ".config", "appstoreconnect", fmt.Sprintf("AuthKey_%s.p8", keyID))
		} else {
			return fmt.Errorf("--private-key or ASC_PRIVATE_KEY environment variable is required")
		}
	}
	return nil
}
