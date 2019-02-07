package docker

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Service contains a docker service definition.
type Service struct {
	Name      string `json:"name,omitempty"`
	Image     string `json:"image,omitempty"`
	Daemon    bool   `json:"daemon,omitempty"`
	Temporary bool   `json:"temporary,omitempty"`

	Command []string `json:"command,omitempty"`

	Environment map[string]string `json:"environment,omitempty"`
	Ports       map[string]int    `json:"ports,omitempty"`
	Volumes     map[string]string `json:"volumes,omitempty"`
}

func (s *Service) isRunning() (bool, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return false, err
	}

	filter := filters.NewArgs()
	filter.Add("name", s.Name)

	containers, err := cli.ContainerList(
		context.Background(), types.ContainerListOptions{Filters: filter},
	)
	if err != nil {
		return false, err
	}

	return len(containers) == 1, nil
}

func (s *Service) actuallyStart() error {
	logrus.Infof("starting service: %s", s.Name)

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	exposedPorts := nat.PortSet{}
	portMapping := nat.PortMap{}

	for hostPort, exposedPort := range s.Ports {
		exPort := nat.Port(fmt.Sprintf("%d/tcp", exposedPort))
		exposedPorts[exPort] = struct{}{}
		portMapping[exPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	var envVars []string

	for varName, varValue := range s.Environment {
		envVars = append(envVars, fmt.Sprintf("%s=%s", varName, varValue))
	}

	// TODO: Take volumes & temporary into account
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        s.Image,
		Cmd:          s.Command,
		Tty:          true,
		ExposedPorts: exposedPorts,
		Env:          envVars,
	}, &container.HostConfig{PortBindings: portMapping}, nil, s.Name)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	logrus.Infof("service [%s] started", s.Name)

	return nil
}

// Start starts the service (if it is not already running).
func (s *Service) Start() error {
	isRunning, err := s.isRunning()
	if err != nil {
		return err
	}

	if !isRunning {
		if err := s.actuallyStart(); err != nil {
			return err
		}

		isRunning, err = s.isRunning()
		if err != nil {
			return err
		}
		if !isRunning {
			return fmt.Errorf("failed to start service: %s", s.Name)
		}
	} else {
		logrus.Infof("service [%s] is already running", s.Name)
	}

	return nil
}