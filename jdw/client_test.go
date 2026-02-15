package jdw

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetVenues(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": [
			{
				"id": 1,
				"name": "Test Venue",
				"address": {
					"town": "Test Town",
					"postcode": "TE1 1ST",
					"location": {
						"latitude": 51.5,
						"longitude": -0.1
					}
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("App-Version") != "1.2.3" {
			t.Errorf("Expected App-Version '1.2.3', got '%s'", r.Header.Get("App-Version"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	venues, err := client.GetVenues()
	if err != nil {
		t.Fatalf("GetVenues failed: %v", err)
	}

	if len(venues) != 1 {
		t.Errorf("Expected 1 venue, got %d", len(venues))
	}

	if venues[0].Name != "Test Venue" {
		t.Errorf("Expected venue name 'Test Venue', got '%s'", venues[0].Name)
	}
}

func TestGetSettings(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": {
			"minVersion": "6.0.0",
			"urls": {
				"terms": "https://example.com/terms"
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	settings, err := client.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}

	if settings.MinVersion != "6.0.0" {
		t.Errorf("Expected minVersion '6.0.0', got '%s'", settings.MinVersion)
	}
}

func TestGetBanners(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": [
			{
				"campaign": "Test Campaign",
				"imageUrl": "https://example.com/banner.jpg"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	banners, err := client.GetBanners()
	if err != nil {
		t.Fatalf("GetBanners failed: %v", err)
	}

	if len(banners) != 1 {
		t.Errorf("Expected 1 banner, got %d", len(banners))
	}

	if banners[0].Campaign != "Test Campaign" {
		t.Errorf("Expected campaign 'Test Campaign', got '%s'", banners[0].Campaign)
	}
}
