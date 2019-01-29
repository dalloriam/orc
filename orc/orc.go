package orc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

type Orc struct {
	cfg Config
}

func New(cfg Config) (*Orc, error) {
	o := &Orc{cfg}
	if err := o.initializePlugins(); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *Orc) registerPlugin(pluginFile string) error {
	rawData, err := ioutil.ReadFile(pluginFile)
	if err != nil {
		return err
	}

	var manifest PluginManifest
	if err := json.Unmarshal(rawData, &manifest); err != nil {
		return err
	}
	fmt.Println(manifest)

	for actionName, command := range manifest.Actions {
		path := fmt.Sprintf("/%s/%s", manifest.Namespace, actionName)
		fmt.Println("Registered action at", path)
		http.HandleFunc(path, command.getHTTPHandler(actionName))
	}

	if manifest.Init.Command != "" {
		fmt.Printf("Executing init command for %s...\n", manifest.Name)
		if _, err := manifest.Init.Execute(nil); err != nil {
			return err
		}
	}

	return nil
}

func (o *Orc) initializePlugins() error {
	files, err := ioutil.ReadDir(o.cfg.PluginsDirectory)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		pluginPath := path.Join(o.cfg.PluginsDirectory, f.Name())
		fmt.Printf("Registering plugin from %s...", pluginPath)
		if err := o.registerPlugin(pluginPath); err != nil {
			return err
		}
	}

	return nil
}
