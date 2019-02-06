package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/dalloriam/orc/docker"
	log "github.com/sirupsen/logrus"
)

type registrarFunc func(string, string, func(string, map[string]interface{}) ([]byte, error))

// Orc is the root orchestrator component.
type Orc struct {
	dockerDirectory string
	pluginDirectory string

	registrar registrarFunc
}

// New initializes the component according to config.
func New(dockerDefinitionsDirectory, pluginDirectory string, actionRegistrar registrarFunc) (*Orc, error) {
	o := &Orc{
		registrar: actionRegistrar,
		dockerDirectory: dockerDefinitionsDirectory,
		pluginDirectory: pluginDirectory,
	}

	if err := o.initModules(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Orc) initModules() error {
	dockerMod, err := docker.NewController(o.dockerDirectory)
	if err != nil {
		return err
	}

	modules := []Module{dockerMod}

	plugins, err := o.loadPlugins()
	if err != nil {
		return err
	}

	modules = append(modules, plugins...)

	for _, mod := range modules {
		n := mod.Name()

		for _, act := range mod.Actions() {
			o.registrar(n, act, mod.Execute)
		}
	}

	return nil
}

func (o *Orc) registerPlugin(pluginFile string) (Module, error) {
	rawData, err := ioutil.ReadFile(pluginFile)
	if err != nil {
		return nil, err
	}

	var manifest PluginManifest
	if err := json.Unmarshal(rawData, &manifest); err != nil {
		return nil, err
	}

	log.Infof("successfully loaded plugin: %s", manifest.Name())

	if manifest.Init.Command != "" {
		log.Infof("executing init command for plugin: %s", manifest.Name())
		if _, err := manifest.Init.Execute(nil); err != nil {
			return nil, err
		}
		log.Infof("successfully initialized plugin: %s", manifest.Name())
	}

	for actionName, action := range manifest.ActionMap {
		action.PluginDir = o.pluginDirectory
		manifest.ActionMap[actionName] = action
	}

	return &manifest, nil
}

func (o *Orc) loadPlugins() ([]Module, error) {
	files, err := ioutil.ReadDir(o.pluginDirectory)
	if err != nil {
		return nil, err
	}

	var modules []Module

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		pluginPath := path.Join(o.pluginDirectory, f.Name())
		log.Infof("found plugin: %s", pluginPath)

		mod, err := o.registerPlugin(pluginPath)

		if err != nil {
			return nil, err
		}

		modules = append(modules, mod)
	}

	return modules, nil
}

func (o *Orc) Serve(host string, port int) error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}
