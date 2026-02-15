package jdw

import "fmt"

// GetVenues fetches the list of all venues.
func (c *Client) GetVenues() ([]Venue, error) {
	var venues []Venue
	err := c.doRequest("GET", "/api/v0.1/venues", nil, &venues)
	return venues, err
}

// GetVenue fetches details for a specific venue by ID.
func (c *Client) GetVenue(id int) (*Venue, error) {
	var venue Venue
	err := c.doRequest("GET", fmt.Sprintf("/api/v0.1/jdw/venues/%d", id), nil, &venue)
	return &venue, err
}
