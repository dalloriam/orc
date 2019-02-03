package orc

// Module represents an abstract task handler.
// Modules can be both a wrapped plugin or an internal module.
type Module interface {
	Name() string
	Actions() []string

	Execute(actionName string, data map[string]interface{}) ([]byte, error)
}
