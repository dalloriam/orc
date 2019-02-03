package orc

type PluginManifest struct {
	PluginName string             `json:"name"`
	Namespace  string             `json:"namespace"`
	ActionMap  map[string]Command `json:"actions"`
	Init       Command            `json:"init"`
}

func (p *PluginManifest) Name() string {
	return p.PluginName
}

func (p *PluginManifest) Actions() []string {
	var actions []string
	for action := range p.ActionMap {
		actions = append(actions, action)
	}
	return actions
}

func (p *PluginManifest) Execute(actionName string, data map[string]interface{}) ([]byte, error) {
	return []byte{}, nil
}
