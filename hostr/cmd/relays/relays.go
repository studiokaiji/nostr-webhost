package relays

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/studiokaiji/nostr-webhost/hostr/cmd/paths"
)

const PATH = ".nostr_relays"

func AddRelay(relayURL string) error {
	dir, err := paths.GetSettingsDirectory()
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, PATH)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(relayURL + "\n")
	return err
}

func RemoveRelay(targetURL string) error {
	dir, err := paths.GetSettingsDirectory()
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, PATH)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		if line != targetURL {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(filePath, []byte(newContent), 0644)

	return err
}

func GetAllRelays() ([]string, error) {
	// Check if relays are provided via environment variable
	envRelays := os.Getenv("RELAY_URLS")
	if envRelays != "" {
		// Parse comma-separated relay URLs from environment variable
		relayList := strings.Split(envRelays, ",")
		cleanedRelays := []string{}
		for _, relay := range relayList {
			trimmed := strings.TrimSpace(relay)
			if len(trimmed) > 0 {
				cleanedRelays = append(cleanedRelays, trimmed)
			}
		}
		if len(cleanedRelays) > 0 {
			return cleanedRelays, nil
		}
	}

	// Fall back to reading from file if environment variable is not set
	dir, err := paths.GetSettingsDirectory()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, PATH)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data := strings.Split(string(content), "\n")
	lines := []string{}
	for _, line := range data {
		if len(line) > 1 {
			lines = append(lines, line)
		}
	}

	return lines, nil
}
