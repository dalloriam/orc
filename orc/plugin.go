package orc

import "fmt"

// PluginManifest represents a plugin declaration.
type PluginManifest struct {
	PluginName string             `json:"name"`
	Namespace  string             `json:"namespace"`
	ActionMap  map[string]Command `json:"actions"`
	Init       Command            `json:"init"`
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
		return action.Execute(data)
	}
	return []byte{}, fmt.Errorf("no such action: %s", actionName)
}
