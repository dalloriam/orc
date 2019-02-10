package task

// StartPayload represents a command payload sent to the task module.
type StartPayload struct {
	TaskName  string   `json:"name" mapstructure:"name"`
	Arguments []string `json:"arguments"`
}
