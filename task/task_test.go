package task_test

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/docker/docker/api/types"
)

type dockerClientMock struct {
	ShouldImagePullFail bool

	ShouldContainerStartFail bool
	ShouldContainerStopFail  bool
}

func (d *dockerClientMock) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	if d.ShouldImagePullFail {
		return nil, errors.New("something terrible")
	}
	return nil, nil
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
