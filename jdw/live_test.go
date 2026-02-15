package jdw

import (
	"os"
	"testing"
)

func TestLiveGetVenues(t *testing.T) {
	if os.Getenv("JDW_LIVE_TESTS") != "true" {
		t.Skip("Skipping live test: JDW_LIVE_TESTS not set to true")
	}

	token := os.Getenv("JDW_TOKEN")
	if token == "" {
		t.Skip("Skipping live test: JDW_TOKEN not set")
	}

	client := NewClient("6.7.1", token, "Go-Live-Test")
	venues, err := client.GetVenues()
	if err != nil {
		t.Fatalf("Live GetVenues failed: %v", err)
	}

	if len(venues) == 0 {
		t.Error("Expected at least one venue from live API")
	}

	t.Logf("Successfully fetched %d venues from live API", len(venues))
}

func TestLiveGetSettings(t *testing.T) {
	if os.Getenv("JDW_LIVE_TESTS") != "true" {
		t.Skip("Skipping live test: JDW_LIVE_TESTS not set to true")
	}

	token := os.Getenv("JDW_TOKEN")
	if token == "" {
		t.Skip("Skipping live test: JDW_TOKEN not set")
	}

	client := NewClient("6.7.1", token, "Go-Live-Test")
	settings, err := client.GetSettings()
	if err != nil {
		t.Fatalf("Live GetSettings failed: %v", err)
	}

	if settings.MinVersion == "" {
		t.Error("Expected MinVersion in live settings response")
	}

	t.Logf("Live settings MinVersion: %s", settings.MinVersion)
}
