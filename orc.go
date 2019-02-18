package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/dalloriam/orc/task"
	"github.com/dalloriam/orc/version"
	log "github.com/sirupsen/logrus"
)

type registrarFunc func(string, string, func(string, map[string]interface{}) ([]byte, error))

// Orc is the root orchestrator component.
type Orc struct {
	taskDirectory   string
	pluginDirectory string

	registrar registrarFunc
}

// New initializes the component according to config.
func New(taskDefinitionDirectory, pluginDirectory string, actionRegistrar registrarFunc) (*Orc, error) {
	log.Infof("[ORC %s @ %s]", version.VERSION, version.GITCOMMIT)
	o := &Orc{
		registrar:       actionRegistrar,
		taskDirectory:   taskDefinitionDirectory,
		pluginDirectory: pluginDirectory,
	}

	if err := o.initModules(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Orc) initModules() error {
	log.Info("looking for modules...")
	taskMod, err := task.NewController(o.taskDirectory, true)
	if err != nil {
		return err
	}

	modules := []Module{taskMod}

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

	log.Infof("module loading complete: %d modules active", len(modules))

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

	if manifest.Init.Command != "" {
		log.Infof("executing init command for plugin: %s", manifest.Name())
		if _, err := manifest.Init.Execute(nil); err != nil {
			return nil, err
		}
	}

	for actionName, action := range manifest.ActionMap {
		action.PluginDir = o.pluginDirectory
		manifest.ActionMap[actionName] = action
	}

	return &manifest, nil
}

func (o *Orc) loadPlugins() ([]Module, error) {
	log.Info("looking for plugins...")
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

	log.Infof("plugin search complete: %d plugins loaded", len(modules))

	return modules, nil
}

// Serve starts the ORC server on the specified host & port.
func (o *Orc) Serve(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Infof("ORC listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}
