package jdw

// Location represents geographic coordinates.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Address represents a physical address.
type Address struct {
	Line1    string   `json:"line1"`
	Line2    *string  `json:"line2"`
	Line3    *string  `json:"line3"`
	Town     string   `json:"town"`
	County   string   `json:"county"`
	Postcode string   `json:"postcode"`
	Location Location `json:"location"`
}

// Venue represents a Wetherspoon pub.
type Venue struct {
	ID        int     `json:"id"`
	VenueRef  int     `json:"venueRef"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Type      string  `json:"type"`
	IsClosed  bool    `json:"isClosed"`
	Address   Address `json:"address"`
	Franchise string  `json:"franchise"`
}

// Settings represents application configuration.
type Settings struct {
	MinVersion string                 `json:"minVersion"`
	Urls       map[string]string      `json:"urls"`
	Features   map[string]interface{} `json:"features"`
}

// Banner represents a promotional banner.
type Banner struct {
	Campaign string `json:"campaign"`
	ImageURL string `json:"imageUrl"`
	URL      string `json:"url"`
}

// APIResponse is the standard wrapper for API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error,omitempty"`
}
