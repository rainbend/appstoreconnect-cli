package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const metadataDir = ".metadata"

type Localization struct {
	Description     string
	Keywords        string
	PromotionalText string
	WhatsNew        string
}

func Write(platform, locale string, loc Localization) error {
	dir := filepath.Join(metadataDir, platform, locale)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	files := map[string]string{
		"description.txt":      loc.Description,
		"keywords.txt":         loc.Keywords,
		"promotional_text.txt": loc.PromotionalText,
		"whats_new.txt":        loc.WhatsNew,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}

	return nil
}

func Read(platform, locale string) (*Localization, error) {
	dir := filepath.Join(metadataDir, platform, locale)

	loc := &Localization{}

	if data, err := os.ReadFile(filepath.Join(dir, "description.txt")); err == nil {
		loc.Description = strings.TrimSpace(string(data))
	}
	if data, err := os.ReadFile(filepath.Join(dir, "keywords.txt")); err == nil {
		loc.Keywords = strings.TrimSpace(string(data))
	}
	if data, err := os.ReadFile(filepath.Join(dir, "promotional_text.txt")); err == nil {
		loc.PromotionalText = strings.TrimSpace(string(data))
	}
	if data, err := os.ReadFile(filepath.Join(dir, "whats_new.txt")); err == nil {
		loc.WhatsNew = strings.TrimSpace(string(data))
	}

	return loc, nil
}

func ListLocales(platform string) ([]string, error) {
	dir := filepath.Join(metadataDir, platform)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading metadata directory %s: %w", dir, err)
	}

	var locales []string
	for _, entry := range entries {
		if entry.IsDir() {
			locales = append(locales, entry.Name())
		}
	}

	return locales, nil
}
