package jdw

// GetSettings fetches the application settings.
func (c *Client) GetSettings() (*Settings, error) {
	var settings Settings
	err := c.doRequest("GET", "/api/v0.1/settings", nil, &settings)
	return &settings, err
}
