package docker

// StartPayload represents a command payload sent to the docker module.
type StartPayload struct {
	ServiceName string   `json:"service_name" mapstructure:"service_name"`
	Arguments   []string `json:"arguments"`
}
