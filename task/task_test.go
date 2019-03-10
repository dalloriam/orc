package task_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/dalloriam/orc/task"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type containerCreateArgs struct {
	name      string
	container *container.Config
	host      *container.HostConfig
	network   *network.NetworkingConfig
}

type dockerClientMock struct {
	ShouldImagePullFail bool

	ShouldContainerCreateFail bool
	ContainerCreateResults    container.ContainerCreateCreatedBody
	createdContainers         []containerCreateArgs

	ShouldContainerInspectFail bool
	ContainerInspectResults    types.ContainerJSON

	ShouldContainerListFail bool
	ContainerListResults    []types.Container

	ShouldContainerRemoveFail bool
	ShouldContainerStartFail  bool
	ShouldContainerStopFail   bool
}

func (d *dockerClientMock) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	if d.ShouldImagePullFail {
		return nil, errors.New("something terrible")
	}
	return nil, nil
}

func (d *dockerClientMock) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	if d.ShouldContainerCreateFail {
		return container.ContainerCreateCreatedBody{}, errors.New("something bad")
	}
	d.createdContainers = append(d.createdContainers, containerCreateArgs{name: containerName, host: hostConfig, container: config, network: networkingConfig})
	return d.ContainerCreateResults, nil
}

func (d *dockerClientMock) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if d.ShouldContainerInspectFail {
		return types.ContainerJSON{}, errors.New("something bad")
	}
	return d.ContainerInspectResults, nil
}

func (d *dockerClientMock) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	if d.ShouldContainerListFail {
		return nil, errors.New("something bad")
	}
	return d.ContainerListResults, nil
}

func (d *dockerClientMock) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	if d.ShouldContainerRemoveFail {
		return errors.New("something bad")
	}
	return nil
}

func (d *dockerClientMock) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	if d.ShouldContainerStartFail {
		return errors.New("something bad")
	}
	return nil
}

func (d *dockerClientMock) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	if d.ShouldContainerStopFail {
		return errors.New("something bad")
	}
	return nil
}

func TestTask_IsRunning(t *testing.T) {
	type testCase struct {
		name string

		listedContainers []types.Container
		listShouldFail   bool

		expectedResult bool
		wantErr        bool
	}

	cases := []testCase{
		testCase{"detects running container", []types.Container{types.Container{}}, false, true, false},
		testCase{"detects non-running container", []types.Container{}, false, false, false},
		testCase{"fails if list() fails", nil, true, false, true},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mockClient := &dockerClientMock{ShouldContainerListFail: tCase.listShouldFail, ContainerListResults: tCase.listedContainers}

			task := task.Task{Client: mockClient}

			isRunning, err := task.IsRunning()

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected error: %v, got err=%v", tCase.wantErr, err)
			}

			if isRunning != tCase.expectedResult {
				t.Errorf("expected isRunning=%v, got %v", tCase.expectedResult, isRunning)
			}
		})
	}
}

func TestTask_Initialize(t *testing.T) {
	type testCase struct {
		name string

		imagePullFails bool
		wantErr        bool
	}

	cases := []testCase{
		{"succeeds if image pull succeeds", false, false},
		{"fails if image pull fails", false, false},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			task := task.Task{Client: &dockerClientMock{ShouldImagePullFail: tCase.imagePullFails}}

			err := task.Initialize()

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected error: %v, got err=%v", tCase.wantErr, err)
			}
		})
	}
}

func TestTask_Start(t *testing.T) {
	type testCase struct {
		name string

		isRunningAfterCreate bool

		task *task.Task

		listFails   bool
		createFails bool
		startFails  bool
		wantErr     bool
	}

	cases := []testCase{
		{
			name:                 "starts simple task successfully",
			isRunningAfterCreate: true,
			task:                 &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"}},
			listFails:            false, createFails: false, startFails: false, wantErr: false,
		},
		{
			name:                 "starts task with volumes successfully",
			isRunningAfterCreate: true,
			task: &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"},
				Volumes: map[string]string{"hello": "world", "test": "volume"}},
			listFails: false, createFails: false, startFails: false, wantErr: false,
		},
		{
			name:                 "starts task with env successfully",
			isRunningAfterCreate: true,
			task: &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"},
				Environment: map[string]string{"hello": "world", "asdf": "asdf"}},
			listFails: false, createFails: false, startFails: false, wantErr: false,
		},
		{
			name:                 "starts task with ports successfully",
			isRunningAfterCreate: true,
			task: &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"},
				Ports: map[string]int{"8080": 8080, "8081": 8081}},
			listFails: false, createFails: false, startFails: false, wantErr: false,
		},
		{
			name:                 "fails if container fails to start",
			isRunningAfterCreate: false,
			task:                 &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"}},
			listFails:            false, createFails: false, startFails: false, wantErr: true,
		},
		{
			name:                 "fails if container listing fails",
			isRunningAfterCreate: true,
			task:                 &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"}},
			listFails:            true, createFails: false, startFails: false, wantErr: true,
		},
		{
			name:                 "fails if container creation fails",
			isRunningAfterCreate: true,
			task:                 &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"}},
			listFails:            false, createFails: true, startFails: false, wantErr: true,
		},
		{
			name:                 "fails if container start fails",
			isRunningAfterCreate: true,
			task:                 &task.Task{Name: "some task", Image: "hello:latest", Daemon: true, Command: []string{"echo", "hello"}},
			listFails:            false, createFails: false, startFails: true, wantErr: true,
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {

			var expectedListResults []types.Container
			if tCase.isRunningAfterCreate {
				expectedListResults = append(expectedListResults, types.Container{})
			}

			mockClient := &dockerClientMock{
				ShouldContainerListFail:   tCase.listFails,
				ShouldContainerCreateFail: tCase.createFails,
				ShouldContainerStartFail:  tCase.startFails,
				ContainerListResults:      expectedListResults,
			}

			tCase.task.Client = mockClient

			err := tCase.task.Start()

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected error: %v, got err=%v", tCase.wantErr, err)
				return
			}

			if err != nil {
				return
			}

			// Additional testing when container started successfully
			if len(mockClient.createdContainers) != 1 {
				t.Errorf("client.CreateContainer() method was not called by task")
				return
			}

			containerDef := mockClient.createdContainers[0]

			if containerDef.name != tCase.task.Name {
				t.Errorf("expected container name=%s, got %s", tCase.task.Name, containerDef.name)
			}

			// Port validation
			// TODO: Validate the formatting & values of the ports.
			if len(containerDef.container.ExposedPorts) != len(tCase.task.Ports) {
				t.Errorf("expected %d exposed ports, got %d instead", len(tCase.task.Ports), len(containerDef.container.ExposedPorts))
			}

			if !(len(containerDef.host.PortBindings) == len(tCase.task.Ports)) {
				t.Errorf("expected %d port mappings, got %d instead", len(tCase.task.Ports), len(containerDef.host.PortBindings))
			}

			// Env var validation
			// TODO: Validate the formatting & values of the env.
			if len(containerDef.container.Env) != len(tCase.task.Environment) {
				t.Errorf("expected %d env var declarations, got %d instead", len(tCase.task.Environment), len(containerDef.host.PortBindings))
			}

			if len(containerDef.host.Binds) != len(tCase.task.Volumes) {
				t.Errorf("expected %d volume declarations, got %d instead", len(tCase.task.Volumes), len(containerDef.host.Binds))
			}
		})
	}
}

func TestTask_Stop(t *testing.T) {
	type testCase struct {
		name string

		listFails        bool
		listedContainers []types.Container

		stopFails bool
		wantErr   bool
	}

	cases := []testCase{
		{"stops containers correctly", false, []types.Container{types.Container{ID: "hello"}}, false, false},
		{"fails when container is not running", false, []types.Container{}, false, true},
		{"fails when list fails", true, nil, false, true},
		{"fails when stop fails", false, []types.Container{types.Container{ID: "hello"}}, true, true},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mockClient := &dockerClientMock{
				ShouldContainerListFail: tCase.listFails,
				ShouldContainerStopFail: tCase.stopFails,
				ContainerListResults:    tCase.listedContainers,
			}

			task := &task.Task{Client: mockClient}

			// TODO: Validate that the container killed by stop() is the correct one.
			err := task.Stop()

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected error: %v, got err=%v", tCase.wantErr, err)
				return
			}
		})
	}
}

func TestTask_Cleanup(t *testing.T) {
	type testCase struct {
		name string

		listFails   bool
		listResults []types.Container

		removeFails bool

		wantErr bool
	}

	cases := []testCase{
		{"cleans up exited container", false, []types.Container{types.Container{}}, false, false},
		{"doesnt fail when container doesnt exists", false, []types.Container{}, false, false},
		{"fails when list fails", true, []types.Container{}, false, true},
		{"fails when remove fails", false, []types.Container{types.Container{}}, true, true},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mockClient := &dockerClientMock{
				ShouldContainerListFail:   tCase.listFails,
				ContainerListResults:      tCase.listResults,
				ShouldContainerRemoveFail: tCase.removeFails,
			}

			task := &task.Task{Client: mockClient}

			err := task.Cleanup()

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected error: %v, got err=%v", tCase.wantErr, err)
				return
			}
			// TODO: Verify that correct container gets cleaned up.
		})
	}
}

func TestTask_NextTasks(t *testing.T) {
	type testCase struct {
		name string

		onSuccess []string
		onFailure []string

		listFails     bool
		listResults   []types.Container
		inspectFails  bool
		inspectReturn types.ContainerJSON

		expectedTasks []string
		wantErr       bool
	}

	succCJSON := types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{}}
	succCJSON.State = &types.ContainerState{ExitCode: 0, Running: false}

	runCJSON := types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{}}
	runCJSON.State = &types.ContainerState{ExitCode: -1, Running: true}

	failCJSON := types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{}}
	failCJSON.State = &types.ContainerState{ExitCode: 1, Running: false}

	cases := []testCase{
		{"succeeds with no next tasks", nil, nil, false, []types.Container{types.Container{}}, false, succCJSON, nil, false},
		{"succeeds with some next tasks", []string{"a", "b"}, nil, false, []types.Container{types.Container{}}, false, succCJSON, nil, false},
		{"succeeds with some next tasks on failure", nil, []string{"a", "b"}, false, []types.Container{types.Container{}}, false, failCJSON, nil, false},
		{"fails when inspect fails", nil, nil, false, []types.Container{types.Container{}}, true, succCJSON, nil, true},
		{"fails when list fails", nil, nil, true, []types.Container{types.Container{}}, false, succCJSON, nil, true},
		{"doesnt fail when container cleaned up", nil, nil, false, []types.Container{}, false, succCJSON, nil, false},
		{"fails when container running", nil, nil, false, []types.Container{types.Container{}}, false, runCJSON, nil, true},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			mockClient := &dockerClientMock{
				ShouldContainerListFail:    tCase.listFails,
				ShouldContainerInspectFail: tCase.inspectFails,
				ContainerInspectResults:    tCase.inspectReturn,
				ContainerListResults:       tCase.listResults,
			}

			task := &task.Task{OnSuccess: tCase.onSuccess, OnFailure: tCase.onFailure, Client: mockClient}
			nexts, err := task.NextTasks()

			if (err != nil) != tCase.wantErr {
				t.Errorf("expected error: %v, got err=%v", tCase.wantErr, err)
				return
			}

			for i := 0; i < len(tCase.expectedTasks); i++ {
				if i >= len(nexts) {
					t.Errorf("not enough tasks returned, expected %d, got %d", len(tCase.expectedTasks), len(nexts))
					return
				}
				if tCase.expectedTasks[i] != nexts[i] {
					t.Errorf("expected task %d to be %s, got %s instead", i, tCase.expectedTasks[i], nexts[i])
				}
			}
		})
	}
}
