package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/KRoperUK/get_spoons/jdw"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "test-value")
	defer os.Unsetenv("TEST_ENV_VAR")

	val := getEnv("TEST_ENV_VAR", "fallback")
	if val != "test-value" {
		t.Errorf("Expected test-value, got %s", val)
	}

	val = getEnv("NON_EXISTENT", "fallback")
	if val != "fallback" {
		t.Errorf("Expected fallback, got %s", val)
	}
}

func TestWriteJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	w := &strings.Builder{}
	err := writeJSON(w, data)
	if err != nil {
		t.Fatalf("writeJSON failed: %v", err)
	}

	expected := "{\n  \"key\": \"value\"\n}\n"
	if w.String() != expected {
		t.Errorf("Expected %q, got %q", expected, w.String())
	}
}

func TestWriteCSV(t *testing.T) {
	venues := []jdw.Venue{
		{
			Name: "Test Pub",
			Address: jdw.Address{
				Line1: "Line 1",
				Town:  "Town",
				Location: jdw.Location{
					Latitude:  51.5,
					Longitude: -0.1,
				},
			},
		},
	}
	w := &strings.Builder{}
	err := writeCSV(w, venues)
	if err != nil {
		t.Fatalf("writeCSV failed: %v", err)
	}

	if !strings.Contains(w.String(), "Test Pub") {
		t.Errorf("Expected Test Pub in output, got %q", w.String())
	}
	if !strings.Contains(w.String(), "51.50000000") {
		t.Errorf("Expected latitude in output, got %q", w.String())
	}
}

func TestSearchVenues(t *testing.T) {
	venues := []jdw.Venue{
		{Name: "The Moon", Address: jdw.Address{Town: "London", Postcode: "E1 6AN"}},
		{Name: "The Sun", Address: jdw.Address{Town: "Manchester", Postcode: "M1 1AE"}},
		{Name: "The Star", Address: jdw.Address{Town: "Bilston", Postcode: "WV14 0EP"}},
	}

	t.Run("FuzzyName", func(t *testing.T) {
		res := searchVenues(venues, "moon", false)
		if len(res) != 1 || res[0].Name != "The Moon" {
			t.Errorf("Expected The Moon, got %v", res)
		}
	})

	t.Run("FuzzyTown", func(t *testing.T) {
		res := searchVenues(venues, "bilston", false)
		if len(res) != 1 || res[0].Address.Town != "Bilston" {
			t.Errorf("Expected Bilston, got %v", res)
		}
	})

	t.Run("SubstringPostcode", func(t *testing.T) {
		res := searchVenues(venues, "wv14", true)
		if len(res) != 1 || res[0].Name != "The Star" {
			t.Errorf("Expected The Star, got %v", res)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		res := searchVenues(venues, "nomatch", false)
		if len(res) != 0 {
			t.Errorf("Expected 0 matches, got %d", len(res))
		}
	})
}

func TestExpandVenues(t *testing.T) {
	// Simple mock server for details
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "data": {"id": 123, "salesAreas": [{"id": 456}]}}`)
	}))
	defer server.Close()

	client := jdw.NewClient("v", "t", "u")
	client.SetBaseURL(server.URL)

	venues := []jdw.Venue{{VenueRef: 123}}
	res := expandVenues(client, venues, 1, false, false)

	if len(res) != 1 {
		t.Errorf("Expected 1 result, got %d", len(res))
	}
	if id, ok := res[0]["id"].(float64); !ok || id != 123 {
		t.Errorf("Expected id 123, got %v", res[0]["id"])
	}

	t.Run("WithMenusAndItems", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "sales-areas") && strings.Contains(r.URL.Path, "menus") && !strings.Contains(r.URL.Path, "789") {
				// GetMenus
				fmt.Fprint(w, `{"success": true, "data": [{"id": 789}]}`)
			} else if strings.Contains(r.URL.Path, "789") {
				// GetMenuItems
				fmt.Fprint(w, `{"success": true, "data": {"id": 789, "items": []}}`)
			} else {
				// GetVenueDetails
				fmt.Fprint(w, `{"success": true, "data": {"id": 123, "salesAreas": [{"id": 456}]}}`)
			}
		}))
		defer server.Close()

		client.SetBaseURL(server.URL)
		venues := []jdw.Venue{{VenueRef: 123}}
		res := expandVenues(client, venues, 1, true, true)

		if len(res) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(res))
		}

		menus, ok := res[0]["menus"].([]interface{})
		if !ok || len(menus) == 0 {
			t.Errorf("Expected menus, got %v", res[0]["menus"])
		}
	})
}

func TestWriteFormattedOutput(t *testing.T) {
	venues := []jdw.Venue{{Name: "Test"}}

	t.Run("JSON", func(t *testing.T) {
		w := &strings.Builder{}
		err := writeFormattedOutput(w, venues, venues, false)
		if err != nil {
			t.Fatalf("writeFormattedOutput failed: %v", err)
		}
		if !strings.Contains(w.String(), "\"name\": \"Test\"") {
			t.Errorf("Expected JSON output, got %q", w.String())
		}
	})

	t.Run("CSV", func(t *testing.T) {
		w := &strings.Builder{}
		err := writeFormattedOutput(w, venues, venues, true)
		if err != nil {
			t.Fatalf("writeFormattedOutput failed: %v", err)
		}
		if !strings.Contains(w.String(), "Pub Name") {
			t.Errorf("Expected CSV header, got %q", w.String())
		}
	})
}

func TestRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "venues") {
			fmt.Fprint(w, `{"success": true, "data": [{"id": 1, "venueRef": 10}]}`)
		} else if strings.Contains(r.URL.Path, "venues/") {
			fmt.Fprint(w, `{"success": true, "data": {"id": 1, "venueRef": 10}}`)
		} else {
			fmt.Fprint(w, `{"success": true, "data": {}}`)
		}
	}))
	defer server.Close()

	os.Setenv("JDW_TOKEN", "test-token")
	defer os.Unsetenv("JDW_TOKEN")

	t.Run("Version", func(t *testing.T) {
		err := Run([]string{"-version"})
		if err != nil {
			t.Errorf("Run -version failed: %v", err)
		}
	})

	t.Run("Help", func(t *testing.T) {
		// Use a non-existent flag to trigger error/usage
		err := Run([]string{"-invalid-flag"})
		if err == nil {
			t.Error("Expected error for invalid flag")
		}
	})

	t.Run("FullRun", func(t *testing.T) {
		os.Setenv("JDW_API_URL", server.URL)
		defer os.Unsetenv("JDW_API_URL")

		// Use a pipe to capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Run([]string{"-limit", "1"})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Run failed: %v", err)
		}

		out, _ := io.ReadAll(r)
		if !strings.Contains(string(out), "\"id\": 1") {
			t.Errorf("Expected id 1 in output, got %s", string(out))
		}
	})

	t.Run("VenueFlag", func(t *testing.T) {
		os.Setenv("JDW_API_URL", server.URL)
		defer os.Unsetenv("JDW_API_URL")

		err := Run([]string{"-venue", "1"})
		if err != nil {
			t.Errorf("Run -venue 1 failed: %v", err)
		}
	})

	t.Run("CSVOutput", func(t *testing.T) {
		os.Setenv("JDW_API_URL", server.URL)
		defer os.Unsetenv("JDW_API_URL")

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Run([]string{"-limit", "1", "-csv"})

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Run -csv failed: %v", err)
		}

		out, _ := io.ReadAll(r)
		if !strings.Contains(string(out), "Pub Name") {
			t.Errorf("Expected CSV header in output, got %s", string(out))
		}
	})
}

func TestFilterVenueForItems(t *testing.T) {
	getVenue := func() map[string]interface{} {
		return map[string]interface{}{
			"name": "Test Pub",
			"menus": []interface{}{
				map[string]interface{}{
					"name": "Drinks",
					"details": map[string]interface{}{
						"sections": []interface{}{
							map[string]interface{}{
								"name": "Beer",
								"items": []interface{}{
									map[string]interface{}{"name": "Stella Artois Pint", "price": 4.5},
									map[string]interface{}{"name": "Peroni", "price": 5.0},
								},
							},
						},
					},
				},
			},
		}
	}

	t.Run("MatchItem", func(t *testing.T) {
		v := getVenue()
		matched := filterVenueForItems(v, "stella pint")
		if !matched {
			t.Fatalf("Expected match for stella pint")
		}
		menus := v["menus"].([]interface{})
		if len(menus) != 1 {
			t.Fatalf("Expected 1 menu, got %d", len(menus))
		}
		m := menus[0].(map[string]interface{})
		details := m["details"].(map[string]interface{})
		sections := details["sections"].([]interface{})
		section := sections[0].(map[string]interface{})
		items := section["items"].([]interface{})
		if len(items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(items))
		}
		itemName := items[0].(map[string]interface{})["name"].(string)
		if itemName != "Stella Artois Pint" {
			t.Errorf("Expected Stella Artois Pint, got %s", itemName)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		v := getVenue()
		matched := filterVenueForItems(v, "guinness")
		if matched {
			t.Errorf("Expected no match for guinness")
		}
	})

	t.Run("EmptyQuery", func(t *testing.T) {
		v := getVenue()
		matched := filterVenueForItems(v, "")
		if !matched {
			t.Errorf("Expected match (no-op) for empty query")
		}
	})
}
