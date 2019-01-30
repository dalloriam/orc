package docker

import (
	"context"
	"fmt"
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
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
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
	fmt.Printf("starting service: %s...\n", s.Name)

	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
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
				HostIP: "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: s.Image,
		Cmd: s.Command,
		Tty: true,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{PortBindings: portMapping}, nil, s.Name)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

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
		fmt.Println("service is already running")
	}

	return nil
}
