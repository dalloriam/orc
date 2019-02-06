package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// Controller defines available docker interactions
type Controller struct {
	defsDirectory string
	services map[string]*Service
}

// NewController loads the service definitions and returns a new controller.
func NewController(definitionsDirectory string) (*Controller, error) {
	cont := &Controller{defsDirectory: definitionsDirectory}
	if err := cont.loadServices(); err != nil {
		return nil, err
	}

	return cont, nil
}

// Name returns the name of the module.
func (c *Controller) Name() string {
	return "docker"
}

// Actions returns the actions defined by the module
func (c *Controller) Actions() []string {
	return []string{"start", "stop"}
}

func (c *Controller) loadServices() error {
	files, err := ioutil.ReadDir(c.defsDirectory)
	if err != nil {
		return fmt.Errorf("invalid docker services directory: %s", c.defsDirectory)
	}

	c.services = make(map[string]*Service)

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
	}
	return json.Marshal(map[string]interface{}{"message": "OK"})
}

// Start starts a service.
func (c *Controller) Start(serviceName string) error {
	if svc, ok := c.services[serviceName]; ok {
		// Start the service from the definition
		if err := svc.Start(); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown service: %s", serviceName)
}
