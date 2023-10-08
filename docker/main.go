// Package main provides an application for processing Docker logs based on YAML configuration and storing them in MongoDB.
package main

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func main() {
	ctx := context.Background()

	cli, err := initializeDockerClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize docker client: %v", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "type=sandbox",
		}),
	})
	if err != nil {
		log.Fatalf("Failed to fetch container list: %v", err)
	}

	mongoClient, collection, err := connectToMongoDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to mongodb: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	for _, container := range containers {
		go processLogs(ctx, cli, collection, container)
	}

	select {}
}
