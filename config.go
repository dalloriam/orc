package main

import (
	"os/user"
	"path"

	"github.com/dalloriam/orc/docker"
	"github.com/spf13/viper"
)

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}

// Config holds the full ORC configuration.
type Config struct {
	Debug            bool          `json:"debug"`
	DockerConfig     docker.Config `json:"docker" mapstructure:"docker"`
	PluginsDirectory string        `json:"plugins_directory" mapstructure:"plugins_directory"`
}

// LoadConfiguration loads & returns the ORC config.
func LoadConfiguration() (Config, error) {
	viper.SetConfigType("json")

	viper.AddConfigPath("./")

	homedir, err := getHomeDir()
	if err != nil {
		return Config{}, err
	}

	viper.AddConfigPath(path.Join(homedir, ".config", "dalloriam"))

	viper.SetConfigName("orc")
	viper.ReadInConfig()

	cfg := Config{}
	viper.Unmarshal(&cfg)

	return cfg, nil
}
