package orc

import (
	"github.com/dalloriam/orc/orc/docker"
)

// Config holds the full ORC configuration.
type Config struct {
	Debug            bool          `json:"debug"`
	DockerConfig     docker.Config `json:"docker" mapstructure:"docker"`
	PluginsDirectory string        `json:"plugins_directory" mapstructure:"plugins_directory"`
}
