package jdw

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const DefaultBaseURL = "https://ca.jdw-apps.net"

// Client is a JDW API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	appVersion string
	token      string
	userAgent  string
	debug      bool
}

// SetDebug enables or disables debug logging for the client.
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// SetBaseURL sets the base URL for the client. This is useful for testing with mock servers.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// GetVenueDetails fetches the full details for a specific venue by ID and returns it as a raw map.
// This is useful for retrieving fields that are not defined in the Venue struct.
func (c *Client) GetVenueDetails(id int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest("GET", fmt.Sprintf("/api/v0.1/jdw/venues/%d", id), nil, &result)
	return result, err
}

// GetMenus fetches the menus for a specific venue and sales area.
func (c *Client) GetMenus(venueID, salesAreaID int) ([]interface{}, error) {
	var result []interface{}
	err := c.doRequest("GET", fmt.Sprintf("/api/v0.1/jdw/venues/%d/sales-areas/%d/menus", venueID, salesAreaID), nil, &result)
	return result, err
}

// GetMenuItems fetches the details (items, sections) for a specific menu.
func (c *Client) GetMenuItems(venueID, salesAreaID, menuID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest("GET", fmt.Sprintf("/api/v0.1/jdw/venues/%d/sales-areas/%d/menus/%d", venueID, salesAreaID, menuID), nil, &result)
	return result, err
}

// NewClient creates a new JDW API client.
func NewClient(appVersion, token, userAgent string) *Client {
	return &Client{
		httpClient: &http.Client{},
		baseURL:    DefaultBaseURL,
		appVersion: appVersion,
		token:      token,
		userAgent:  userAgent,
	}
}

func (c *Client) doRequest(method, path string, body io.Reader, result any) (err error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}

	if c.debug {
		fmt.Printf("DEBUG: %s %s\n", method, c.baseURL+path)
	}

	req.Header.Set("App-Version", c.appVersion)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %v", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// We use a generic response wrapper to unmarshal accurately
	wrapper := struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}{}

	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		return err
	}

	if !wrapper.Success {
		return fmt.Errorf("API response indicated failure")
	}

	return json.Unmarshal(wrapper.Data, result)
}
