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

func (c *Client) doRequest(method, path string, body io.Reader, result any) error {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("App-Version", c.appVersion)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
