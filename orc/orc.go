package orc

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"

	"github.com/dalloriam/orc/orc/docker"

	"go.uber.org/zap"
)

// Orc is the root orchestrator component.
type Orc struct {
	cfg       Config
	log       *zap.SugaredLogger
	registrar func(moduleName, actionName string, fn func(actionName string, data map[string]interface{}) ([]byte, error))
}

// New initializes the component according to config.
func New(cfg Config, actionResgistrar func(moduleName, actionName string, fn func(actionName string, data map[string]interface{}) ([]byte, error))) (*Orc, error) {
	var logger *zap.Logger
	var err error
	if cfg.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}

	defer logger.Sync()

	sugared := logger.Sugar()

	o := &Orc{
		cfg:       cfg,
		log:       sugared,
		registrar: actionResgistrar,
	}

	o.log.Infof("configuration loaded")

	if err := o.initModules(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Orc) initModules() error {
	dockerMod, err := docker.NewController(o.cfg.DockerConfig)
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

	o.log.Infof("successfully loaded plugin: %s", manifest.Name())

	if manifest.Init.Command != "" {
		o.log.Infof("executing init command for plugin: %s", manifest.Name())
		if _, err := manifest.Init.Execute(nil); err != nil {
			return nil, err
		}
		o.log.Infof("successfully initialized plugin: %s", manifest.Name())
	}

	return &manifest, nil
}

func (o *Orc) loadPlugins() ([]Module, error) {
	files, err := ioutil.ReadDir(o.cfg.PluginsDirectory)
	if err != nil {
		return nil, err
	}

	var modules []Module

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		pluginPath := path.Join(o.cfg.PluginsDirectory, f.Name())
		o.log.Infof("found plugin: %s", pluginPath)

		mod, err := o.registerPlugin(pluginPath)

		if err != nil {
			return nil, err
		}

		modules = append(modules, mod)
	}

	return modules, nil
}
