package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
	ctx context.Context
}

type Container struct {
	ID   string
	Name string
}

// creating a new docker client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker: %w", err)
	}

	return &Client{
		cli: cli,
		ctx: context.Background(),
	}, nil

}

// Ping: checks if docker daemon responds
func (c *Client) Ping() error {
	_, err := c.cli.Ping(c.ctx)
	if err != nil {
		return fmt.Errorf("docker daemon not responding: %w", err)
	}
	return nil
}

// get docker version
func (c *Client) GetVersion() (types.Version, error) {
	v, err := c.cli.ServerVersion(c.ctx)
	if err != nil {
		return types.Version{}, fmt.Errorf("can't get docker version: %w", err)
	}
	return v, nil
}

// list container
func (c *Client) ListContainers() ([]Container, error) {
	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []Container
	for _, ctr := range containers {
		name := "unknown"
		if len(ctr.Names) > 0 {
			name = ctr.Names[0][1:]
		}

		result = append(result, Container{
			ID:   ctr.ID[:12],
			Name: name,
		})
	}

	return result, nil
}

// close client connection
func (c *Client) Close() error {
	return c.cli.Close()
}
