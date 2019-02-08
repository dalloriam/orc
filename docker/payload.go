package docker

// StartPayload represents a command payload sent to the docker module.
type StartPayload struct {
	ServiceName string   `json:"name" mapstructure:"name"`
	Arguments   []string `json:"arguments"`
}
