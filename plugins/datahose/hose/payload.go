package hose

type Payload struct {
	Key  string                 `json:"key"`
	Body map[string]interface{} `json:"body"`
}
