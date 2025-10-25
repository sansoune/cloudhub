package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
	ctx context.Context
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

// close client connection
func (c *Client) Close() error {
	return c.cli.Close()
}
