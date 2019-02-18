package management

import (
	"encoding/json"
	"fmt"
)

const (
	managementModName = "manage"

	getActionsAction = "actions_available"
)

// Module manages an ORC instance.
type Module struct {
	actionMap map[string][]string
}

// NewModule returns a new management module.
func NewModule() *Module {
	return &Module{
		actionMap: make(map[string][]string),
	}
}

// Name returns the name of the management module.
func (m *Module) Name() string { return managementModName }

// Actions returns the actions defined by the management module.
func (m *Module) Actions() []string {
	return []string{getActionsAction}
}

func (m *Module) getActions() ([]byte, error) {
	return json.Marshal(m.actionMap)
}

// RegisterAction adds the action to the manager.
func (m *Module) RegisterAction(moduleName, action string) {
	if _, ok := m.actionMap[moduleName]; !ok {
		m.actionMap[moduleName] = []string{}
	}

	m.actionMap[moduleName] = append(m.actionMap[moduleName], action)
}

// Execute executes a management action.
func (m *Module) Execute(actionName string, data map[string]interface{}) ([]byte, error) {
	switch actionName {
	case getActionsAction:
		return m.getActions()
	}
	return nil, fmt.Errorf("unknown action: %s", actionName)
}
