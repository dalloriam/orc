package orc

type PluginManifest struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Actions   map[string]Command `json:"actions"`
	Init      Command            `json:"init"`
}
