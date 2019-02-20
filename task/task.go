package task

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type taskDef interface {
	IsRunning() (bool, error)
	Start() error
	Stop() error
	Cleanup() error

	NextTasks() ([]string, error)
}

type dockerClient interface {
	ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error)

	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
}

// Task contains a docker task definition.
type Task struct {
	Name   string `json:"name,omitempty"`
	Image  string `json:"image,omitempty"`
	Daemon bool   `json:"daemon,omitempty"`

	Command []string `json:"command,omitempty"`

	Environment map[string]string `json:"environment,omitempty"`
	Ports       map[string]int    `json:"ports,omitempty"`
	Volumes     map[string]string `json:"volumes,omitempty"`

	OnSuccess []string `json:"on_success,omitempty"`
	OnFailure []string `json:"on_failure,omitempty"`

	Client dockerClient
}

func (s *Task) initClient() (dockerClient, error) {
	if s.Client != nil {
		return s.Client, nil
	}

	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	s.Client = client
	return client, nil
}

func (s *Task) containerID() (string, error) {
	cli, err := s.initClient()
	if err != nil {
		return "", err
	}

	filter := filters.NewArgs()
	filter.Add("name", s.Name)

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: filter, All: true})
	if err != nil {
		return "", err
	}

	if len(containers) != 1 {
		return "", fmt.Errorf("unexpected state: %d containers found", len(containers))
	}

	return containers[0].ID, nil
}

// IsRunning returns whether the service is currently running.
func (s *Task) IsRunning() (bool, error) {
	cli, err := s.initClient()
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

// Initialize pulls the docker image associated with the service, if required.
func (s *Task) Initialize() error {
	logrus.Debugf("ensuring image [%s] is available...", s.Image)
	cli, err := s.initClient()
	if err != nil {
		return err
	}

	_, err = cli.ImagePull(context.Background(), s.Image, types.ImagePullOptions{})

	if err == nil {
		logrus.Debugf("image [%s] is available", s.Image)
	}
	return err
}

func (s *Task) actuallyStart() error {
	logrus.Infof("starting service: %s", s.Name)

	cli, err := s.initClient()
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

	var volumeBinds []string

	for srcVol, dstVol := range s.Volumes {
		volumeBinds = append(volumeBinds, fmt.Sprintf("%s:%s", srcVol, dstVol))
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        s.Image,
		Cmd:          s.Command,
		Tty:          true,
		ExposedPorts: exposedPorts,
		Env:          envVars,
	}, &container.HostConfig{
		PortBindings: portMapping,
		Binds:        volumeBinds,
	}, nil, s.Name)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	logrus.Infof("service [%s] started", s.Name)

	return nil
}

// Start starts the service (if it is not already running).
func (s *Task) Start() error {
	if err := s.actuallyStart(); err != nil {
		return err
	}

	isRunning, err := s.IsRunning()
	if err != nil {
		return err
	}
	if !isRunning {
		return fmt.Errorf("failed to start service: %s", s.Name)
	}

	return nil
}

// Stop stops the service.
func (s *Task) Stop() error {
	logrus.Infof("stopping service: %s", s.Name)

	cli, err := s.initClient()
	if err != nil {
		return err
	}

	containerID, err := s.containerID()
	if err != nil {
		return err
	}

	duration := 10 * time.Second
	return cli.ContainerStop(context.Background(), containerID, &duration)
}

// Cleanup picks up the pieces & deletes the container.
func (s *Task) Cleanup() error {

	cli, err := s.initClient()
	if err != nil {
		return err
	}

	containerID, err := s.containerID()
	if err != nil {
		if strings.HasPrefix(err.Error(), "unexpected state") {
			// Container was already cleaned up.
			return nil
		}
		return err
	}

	logrus.Infof("cleaning up container: %s", s.Name)
	return cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
}

// NextTasks fetches the exit status of the container, and
// returns s.OnSuccess if 0, else s.OnFailure.
func (s *Task) NextTasks() ([]string, error) {
	ctxLog := logrus.WithFields(logrus.Fields{
		"module": moduleName,
		"task":   s.Name,
	})

	containerID, err := s.containerID()
	if err != nil {
		if strings.HasPrefix(err.Error(), "unexpected state") {
			// Container was already cleaned up, we can't risk starting anymore tasks.
			ctxLog.Warnf("container was already cleaned up, not risking creation of subsequent tasks")
			return nil, nil
		}
		return nil, err
	}

	cli, err := s.initClient()
	if err != nil {
		return nil, err
	}

	containerInfo, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, err
	}

	if containerInfo.State.Running {
		return nil, fmt.Errorf("container is running")
	}

	ctxLog.Infof("task exited with exit code %d", containerInfo.State.ExitCode)
	if containerInfo.State.ExitCode == 0 {
		return s.OnSuccess, nil
	}

	return s.OnFailure, nil
}
