package plugins

import (
	"encoding/json"
	"fmt"
)

// PluginManifest represents a plugin declaration.
type PluginManifest struct {
	PluginName string             `json:"name"`
	ActionMap  map[string]Command `json:"actions"`
	Init       Command            `json:"init,omitempty"`
}

// Name returns the name of the plugin.
func (p *PluginManifest) Name() string {
	return p.PluginName
}

// Actions returns the actions defined by the plugin.
func (p *PluginManifest) Actions() []string {
	var actions []string
	for action := range p.ActionMap {
		actions = append(actions, action)
	}
	return actions
}

// Execute executes the plugin.
func (p *PluginManifest) Execute(actionName string, data map[string]interface{}) ([]byte, error) {
	if action, ok := p.ActionMap[actionName]; ok {
		output, err := action.Execute(data)
		if err != nil {
			return nil, err
		}

		response := make(map[string]interface{})
		response["output"] = output

		marshalledBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		return marshalledBytes, nil
	}
	return []byte{}, fmt.Errorf("no such action: %s", actionName)
}
