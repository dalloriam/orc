package docker

import (
	"context"
	"fmt"

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
}

// Start starts the service (if it is not already running).
func (s *Service) Start() error {
	isRunning, err := s.isRunning()
	if err != nil {
		return err
	}

	if !isRunning {
		return s.actuallyStart()
	} else {
		fmt.Println("service is already running")
		return nil
	}
}
