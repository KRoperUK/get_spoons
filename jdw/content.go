package jdw

// GetBanners fetches the promotional banners.
func (c *Client) GetBanners() ([]Banner, error) {
	var banners []Banner
	err := c.doRequest("GET", "/api/v0.1/content/promotional-banners", nil, &banners)
	return banners, err
}
