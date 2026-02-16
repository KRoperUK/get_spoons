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
		err := writeFormattedOutput(w, venues, venues, false, false)
		if err != nil {
			t.Fatalf("writeFormattedOutput failed: %v", err)
		}
		if !strings.Contains(w.String(), "\"name\": \"Test\"") {
			t.Errorf("Expected JSON output, got %q", w.String())
		}
	})

	t.Run("CSV", func(t *testing.T) {
		w := &strings.Builder{}
		err := writeFormattedOutput(w, venues, venues, true, false)
		if err != nil {
			t.Fatalf("writeFormattedOutput failed: %v", err)
		}
		if !strings.Contains(w.String(), "Pub Name") {
			t.Errorf("Expected CSV header, got %q", w.String())
		}
	})

	t.Run("YAML", func(t *testing.T) {
		w := &strings.Builder{}
		err := writeFormattedOutput(w, venues, venues, false, true)
		if err != nil {
			t.Fatalf("writeFormattedOutput failed: %v", err)
		}
		if !strings.Contains(w.String(), "name: Test") {
			t.Errorf("Expected YAML output, got %q", w.String())
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

	t.Run("ItemSearch", func(t *testing.T) {
		// Mock server for item search
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.HasSuffix(r.URL.Path, "/venues") {
				// GetVenues
				fmt.Fprint(w, `{"success": true, "data": [{"id": 1, "venueRef": 10, "name": "The Test Pub"}]}`)
			} else if strings.Contains(r.URL.Path, "/venues/10/sales-areas/456/menus/789") {
				// GetMenuItems
				fmt.Fprint(w, `{"success": true, "data": {"id": 789, "items": [{"name": "Burger", "price": 10.0}, {"name": "Salad", "price": 8.0}]}}`)
			} else if strings.Contains(r.URL.Path, "/venues/10/sales-areas/456/menus") {
				// GetMenus
				fmt.Fprint(w, `{"success": true, "data": [{"id": 789, "name": "Food Menu"}]}`)
			} else if strings.Contains(r.URL.Path, "/venues/10") {
				// GetVenueDetails
				fmt.Fprint(w, `{"success": true, "data": {"id": 1, "venueRef": 10, "salesAreas": [{"id": 456}]}}`)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}))
		defer server.Close()

		os.Setenv("JDW_API_URL", server.URL)
		defer os.Unsetenv("JDW_API_URL")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Capture stderr
		oldStderr := os.Stderr
		rErr, wErr, _ := os.Pipe()
		os.Stderr = wErr

		err := Run([]string{"-item-search", "burger"})

		w.Close()
		os.Stdout = oldStdout
		wErr.Close()
		os.Stderr = oldStderr

		if err != nil {
			t.Fatalf("Run -item-search failed: %v", err)
		}

		outStub, _ := io.ReadAll(r)
		outStr := string(outStub)
		errStub, _ := io.ReadAll(rErr)
		errStr := string(errStub)

		// Check output JSON contains "Burger" but NOT "Salad"
		if !strings.Contains(outStr, "Burger") {
			t.Errorf("Expected output to contain 'Burger', got: %s", outStr)
		}
		if strings.Contains(outStr, "Salad") {
			t.Errorf("Expected output NOT to contain 'Salad', got: %s", outStr)
		}

		// Check stderr for filtering message
		if !strings.Contains(errStr, "Filtered results for items matching \"burger\"") {
			t.Errorf("Expected stderr to confirm filtering, got: %s", errStr)
		}
	})

	t.Run("ItemSearch_MultipleVenues", func(t *testing.T) {
		// Mock server returning multiple venues
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.HasSuffix(r.URL.Path, "venues") {
				// GetVenues - Return 2 venues
				fmt.Fprint(w, `{"success": true, "data": [{"id": 1, "venueRef": 10, "name": "Pub One"}, {"id": 2, "venueRef": 20, "name": "Pub Two"}]}`)
			} else if strings.Contains(r.URL.Path, "789") {
				// GetMenuItems for Pub One
				fmt.Fprint(w, `{"success": true, "data": {"id": 789, "items": [{"name": "Burger", "price": 10.0}]}}`)
			} else if strings.Contains(r.URL.Path, "menus") {
				// GetMenus for Pub One
				fmt.Fprint(w, `{"success": true, "data": [{"id": 789, "name": "Food Menu"}]}`)
			} else if strings.Contains(r.URL.Path, "venues/10") {
				// GetVenueDetails for Pub One
				fmt.Fprint(w, `{"success": true, "data": {"id": 1, "venueRef": 10, "salesAreas": [{"id": 456}]}}`)
			} else {
				// Should not be called for Pub Two
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}))
		defer server.Close()

		os.Setenv("JDW_API_URL", server.URL)
		defer os.Unsetenv("JDW_API_URL")

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		oldStderr := os.Stderr
		rErr, wErr, _ := os.Pipe()
		os.Stderr = wErr

		// Use -item-search "burger"
		err := Run([]string{"-item-search", "burger"})

		w.Close()
		os.Stdout = oldStdout
		wErr.Close()
		os.Stderr = oldStderr

		// We need to read from r to clear the buffer, even if unused, to avoid pipe issues
		// However, the error is that it's unused. Let's just discard it.
		_, _ = io.Copy(io.Discard, r)

		if err != nil {
			t.Fatalf("Run -item-search failed: %v", err)
		}

		errBytes, _ := io.ReadAll(rErr)
		errStr := string(errBytes)

		// Check for the warning message
		if !strings.Contains(errStr, "Item search is only allowed on an individual venue. Using the first match") {
			t.Errorf("Expected multiple venue warning, got: %s", errStr)
		}
	})

	t.Run("ItemSearch_NoMatches", func(t *testing.T) {
		// Mock server for item search with no matches
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.HasSuffix(r.URL.Path, "venues") {
				fmt.Fprint(w, `{"success": true, "data": [{"id": 1, "venueRef": 10, "name": "The Test Pub"}]}`)
			} else if strings.Contains(r.URL.Path, "789") {
				fmt.Fprint(w, `{"success": true, "data": {"id": 789, "items": [{"name": "Salad", "price": 8.0}]}}`)
			} else if strings.Contains(r.URL.Path, "menus") {
				fmt.Fprint(w, `{"success": true, "data": [{"id": 789, "name": "Food Menu"}]}`)
			} else if strings.Contains(r.URL.Path, "venues/10") {
				fmt.Fprint(w, `{"success": true, "data": {"id": 1, "venueRef": 10, "salesAreas": [{"id": 456}]}}`)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		}))
		defer server.Close()

		os.Setenv("JDW_API_URL", server.URL)
		defer os.Unsetenv("JDW_API_URL")

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		oldStderr := os.Stderr
		rErr, wErr, _ := os.Pipe()
		os.Stderr = wErr

		err := Run([]string{"-item-search", "steak"})

		w.Close()
		os.Stdout = oldStdout
		wErr.Close()
		os.Stderr = oldStderr

		_, _ = io.Copy(io.Discard, r)

		if err != nil {
			t.Fatalf("Run -item-search failed: %v", err)
		}

		errBytes, _ := io.ReadAll(rErr)
		errStr := string(errBytes)

		if !strings.Contains(errStr, "No items matching \"steak\" found") {
			t.Errorf("Expected no items found message, got: %s", errStr)
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

	t.Run("MatchParentKeepsChildren", func(t *testing.T) {
		// Test that searching for a parent item (e.g. "Burger") keeps its options
		v := map[string]interface{}{
			"menus": []interface{}{
				map[string]interface{}{
					"details": map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{
								"name": "Burger",
								"options": []interface{}{
									map[string]interface{}{"label": "Cheese"},
									map[string]interface{}{"label": "Bacon"},
								},
							},
						},
					},
				},
			},
		}

		matched := filterVenueForItems(v, "burger")
		if !matched {
			t.Fatalf("Expected match for burger")
		}

		menus := v["menus"].([]interface{})
		details := menus[0].(map[string]interface{})["details"].(map[string]interface{})
		items := details["items"].([]interface{})
		burger := items[0].(map[string]interface{})
		options := burger["options"].([]interface{})
		if len(options) != 2 {
			t.Errorf("Expected 2 options strings to be kept when parent matches, got %d", len(options))
		}
	})

	t.Run("MatchChildKeepsParent", func(t *testing.T) {
		// Test that searching for a child option (e.g. "Cheese") keeps the parent item but prunes non-matching siblings
		v := map[string]interface{}{
			"menus": []interface{}{
				map[string]interface{}{
					"details": map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{
								"name": "Burger",
								"options": []interface{}{
									map[string]interface{}{"label": "Cheese"}, // Match
									map[string]interface{}{"label": "Bacon"},  // No match
								},
							},
						},
					},
				},
			},
		}

		matched := filterVenueForItems(v, "cheese")
		if !matched {
			t.Fatalf("Expected match for cheese")
		}

		menus := v["menus"].([]interface{})
		details := menus[0].(map[string]interface{})["details"].(map[string]interface{})
		items := details["items"].([]interface{})
		burger := items[0].(map[string]interface{})
		options := burger["options"].([]interface{})
		if len(options) != 1 {
			t.Errorf("Expected 1 option to be kept when child matches, got %d", len(options))
		} else {
			if options[0].(map[string]interface{})["label"] != "Cheese" {
				t.Errorf("Expected Cheese option, got %v", options[0])
			}
		}
	})
}
