package docker

import (
	"fmt"

	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

// creating a new docker client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker: %w", err)
	}

	return &Client{
		cli: cli,
	}, nil

}

//close client connection
func (c *Client) Close() error {
	return c.cli.Close()
}
