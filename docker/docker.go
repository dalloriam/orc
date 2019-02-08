package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

const (
	moduleName = "service"
)

// Controller defines available docker interactions
type Controller struct {
	defsDirectory string
	services      map[string]serviceDef
}

// NewController loads the service definitions and returns a new controller.
func NewController(definitionsDirectory string) (*Controller, error) {
	cont := &Controller{defsDirectory: definitionsDirectory}
	if err := cont.loadServices(); err != nil {
		return nil, err
	}

	logrus.Infof("%s module loaded successfully", moduleName)
	return cont, nil
}

// Name returns the name of the module.
func (c *Controller) Name() string {
	return moduleName
}

// Actions returns the actions defined by the module
func (c *Controller) Actions() []string {
	return []string{"start", "stop"}
}

func (c *Controller) loadServices() error {
	ctxLog := logrus.WithFields(logrus.Fields{
		"module": moduleName,
	})

	files, err := ioutil.ReadDir(c.defsDirectory)
	if err != nil {
		return fmt.Errorf("invalid services directory: %s", c.defsDirectory)
	}

	c.services = make(map[string]serviceDef)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		servicePath := path.Join(c.defsDirectory, f.Name())

		data, err := ioutil.ReadFile(servicePath)
		if err != nil {
			return err
		}

		var svc Service
		if err := json.Unmarshal(data, &svc); err != nil {
			return err
		}

		c.services[svc.Name] = &svc
		if err := svc.Initialize(); err != nil {
			panic(err)
		}
		ctxLog.Infof("service loaded successfully: %s", svc.Name)
	}

	return nil
}

// Execute executes an action.
func (c *Controller) Execute(actionName string, data map[string]interface{}) ([]byte, error) {
	switch actionName {
	case "start":
		var args StartPayload
		if err := mapstructure.Decode(data, &args); err != nil {
			return nil, err
		}
		if err := c.Start(args.ServiceName); err != nil {
			return nil, err
		}
	case "stop":
		var args StartPayload
		if err := mapstructure.Decode(data, &args); err != nil {
			return nil, err
		}
		if err := c.Stop(args.ServiceName); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown action: %s", actionName)
	}
	return json.Marshal(map[string]interface{}{"message": "OK"})
}

// Start starts a service.
func (c *Controller) Start(serviceName string) error {
	if svc, ok := c.services[serviceName]; ok {
		// Start the service from the definition
		isRunning, err := svc.IsRunning()

		if err != nil {
			return err
		}

		if !isRunning {
			if err := svc.Start(); err != nil {
				return err
			}
		} else {
			logrus.Infof("service [%s] is already running", serviceName)
		}
		return nil
	}

	return fmt.Errorf("unknown service: %s", serviceName)
}

// Stop stops a service.
func (c *Controller) Stop(serviceName string) error {
	if svc, ok := c.services[serviceName]; ok {
		isRunning, err := svc.IsRunning()

		if err != nil {
			return err
		}

		if !isRunning {
			logrus.Infof("service [%s] is not running", serviceName)
			return nil
		}

		// Stop the service from the definition
		return svc.Stop()
	}

	return fmt.Errorf("unknown service: %s", serviceName)
}
