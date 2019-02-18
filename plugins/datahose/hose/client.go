package hose

// Client represents a datahose client.
type Client struct {
	cfg config
}

// NewClient returns a datahose client
func NewClient() (*Client, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, err
	}

	return &Client{cfg}, nil
}

// Push pushes an event to the datahose.
func (c *Client) Push(eventKey string, body map[string]interface{}) error {
	return nil
}
