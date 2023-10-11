package main

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// initializeDockerClient initializes a new Docker client.
func initializeDockerClient(ctx context.Context) (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(ctx)
	return cli, nil
}

// getContainerName retrieves the name of the container.
func getContainerName(container types.Container) string {
	return strings.TrimPrefix(container.Names[0], "/")
}
