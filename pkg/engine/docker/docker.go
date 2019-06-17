package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types/filters"

	"github.com/deviceplane/deviceplane/pkg/engine"
	"github.com/deviceplane/deviceplane/pkg/spec"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var _ engine.Engine = &Engine{}

type Engine struct {
	client *client.Client
}

func NewEngine() (*Engine, error) {
	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &Engine{
		client: client,
	}, nil
}

func (e *Engine) CreateContainer(ctx context.Context, name string, s spec.Service) (string, error) {
	config, hostConfig, err := convert(s)
	if err != nil {
		return "", err
	}

	resp, err := e.client.ContainerCreate(ctx, config, hostConfig, nil, name)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (e *Engine) StartContainer(ctx context.Context, id string) error {
	if err := e.client.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		// TODO
		if strings.Contains(err.Error(), "No such container") {
			return engine.ErrInstanceNotFound
		}
		return err
	}
	return nil
}

func (e *Engine) ListContainers(ctx context.Context, keyFilters map[string]struct{}, keyAndValueFilters map[string]string, all bool) ([]engine.Instance, error) {
	args := filters.NewArgs()
	for k := range keyFilters {
		args.Add("label", k)
	}
	for k, v := range keyAndValueFilters {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := e.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     all,
	})
	if err != nil {
		return nil, err
	}

	var instances []engine.Instance
	for _, container := range containers {
		instances = append(instances, convertToInstance(container))
	}

	return instances, nil
}

func (e *Engine) StopContainer(ctx context.Context, id string) error {
	if err := e.client.ContainerStop(ctx, id, nil); err != nil {
		// TODO
		if strings.Contains(err.Error(), "No such container") {
			return engine.ErrInstanceNotFound
		}
		return engine.ErrInstanceNotFound
	}
	return nil
}

func (e *Engine) RemoveContainer(ctx context.Context, id string) error {
	if err := e.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		// TODO
		if strings.Contains(err.Error(), "No such container") {
			return engine.ErrInstanceNotFound
		}
		return engine.ErrInstanceNotFound
	}
	return nil
}

func (e *Engine) PullImage(ctx context.Context, image string) error {
	out, err := e.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(ioutil.Discard, out)
	return err
}
