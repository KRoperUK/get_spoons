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
		if _, err := fmt.Fprint(w, mockResponse); err != nil {
			t.Errorf("failed to write mock response: %v", err)
		}
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
		if _, err := fmt.Fprint(w, mockResponse); err != nil {
			t.Errorf("failed to write mock response: %v", err)
		}
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
		if _, err := fmt.Fprint(w, mockResponse); err != nil {
			t.Errorf("failed to write mock response: %v", err)
		}
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

func TestGetVenueDetails(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": {
			"id": 123,
			"name": "Detailed Venue",
			"extraField": "extraValue"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, mockResponse); err != nil {
			t.Errorf("failed to write mock response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	details, err := client.GetVenueDetails(123)
	if err != nil {
		t.Fatalf("GetVenueDetails failed: %v", err)
	}

	if name, ok := details["name"].(string); !ok || name != "Detailed Venue" {
		t.Errorf("Expected name 'Detailed Venue', got '%v'", details["name"])
	}

	if extra, ok := details["extraField"].(string); !ok || extra != "extraValue" {
		t.Errorf("Expected extraField 'extraValue', got '%v'", details["extraField"])
	}
}

func TestGetVenue(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": {
			"id": 1,
			"name": "Single Venue"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	venue, err := client.GetVenue(1)
	if err != nil {
		t.Fatalf("GetVenue failed: %v", err)
	}

	if venue.Name != "Single Venue" {
		t.Errorf("Expected name 'Single Venue', got '%s'", venue.Name)
	}
}

func TestGetMenus(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": [
			{"id": 10, "name": "Main Menu"}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	menus, err := client.GetMenus(123, 456)
	if err != nil {
		t.Fatalf("GetMenus failed: %v", err)
	}

	if len(menus) != 1 {
		t.Errorf("Expected 1 menu, got %d", len(menus))
	}
}

func TestGetMenuItems(t *testing.T) {
	mockResponse := `{
		"success": true,
		"data": {
			"id": 789,
			"categories": [
				{"id": 1000, "name": "Burgers"}
			]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	client := NewClient("1.2.3", "test-token", "test-ua")
	client.baseURL = server.URL
	details, err := client.GetMenuItems(123, 456, 789)
	if err != nil {
		t.Fatalf("GetMenuItems failed: %v", err)
	}

	if id, ok := details["id"].(float64); !ok || id != 789 {
		t.Errorf("Expected id 789, got %v", details["id"])
	}
}

func TestSetDebug(t *testing.T) {
	client := NewClient("1.2.3", "token", "ua")
	client.SetDebug(true)
	if !client.debug {
		t.Error("Expected debug to be true")
	}
}

func TestDoRequestErrors(t *testing.T) {
	t.Run("StatusNotOK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := NewClient("v", "t", "u")
		client.baseURL = server.URL
		_, err := client.GetVenues()
		if err == nil || err.Error() != "API returned status: 403 Forbidden" {
			t.Errorf("Expected 403 error, got %v", err)
		}
	})

	t.Run("SuccessFalse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"success": false}`)
		}))
		defer server.Close()

		client := NewClient("v", "t", "u")
		client.baseURL = server.URL
		_, err := client.GetVenues()
		if err == nil || err.Error() != "API response indicated failure" {
			t.Errorf("Expected failure error, got %v", err)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `invalid`)
		}))
		defer server.Close()

		client := NewClient("v", "t", "u")
		client.baseURL = server.URL
		_, err := client.GetVenues()
		if err == nil {
			t.Error("Expected JSON unmarshal error")
		}
	})
}
