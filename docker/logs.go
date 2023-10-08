// Package main provides functionalities for processing and storing logs.
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogEntry struct {
	ExecutionID   string `json:"execution_id"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}

// LogData contains the log data to be stored.
type LogData struct {
	Container    string      `json:"container"`
	FriendlyName string      `json:"friendly_name"`
	Timestamp    string      `json:"timestamp"`
	Target       string      `json:"target"`
	Log          LogEntry    `json:"log"`
	Error        interface{} `json:"error,omitempty"`
}

// getLogHash computes a hash based on the container ID and log line.
func getLogHash(containerID, logLine string) string {
	data := containerID + logLine
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// processLogs fetches and processes logs for the specified container.
func processLogs(ctx context.Context, cli *client.Client, collection *mongo.Collection, container types.Container) {
	containerName := getContainerName(container)

	reader, err := cli.ContainerLogs(ctx, container.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
	if err != nil {
		log.Printf("Failed to fetch logs for container %s: %v", container.ID, err)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		handleLogEntry(ctx, collection, scanner.Text(), container.ID, containerName)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error for container %s: %v", container.ID, err)
	}

}

// handleLogEntry processes and stores a log line.
func handleLogEntry(ctx context.Context, collection *mongo.Collection, logLineStr, containerID, containerName string) {
	var parsedLog map[string]interface{}

	logLineStr = logLineStr[8:]

	if err := json.Unmarshal([]byte(logLineStr), &parsedLog); err != nil {
		log.Println("Failed to parse log. Skipping.")
		return
	}

	timestamp, _ := parsedLog["timestamp"].(string)
	target, _ := parsedLog["target"].(string)
	logError := parsedLog["error"]

	logEntry := LogEntry{
		ExecutionID:   parsedLog["execution_id"].(string),
		CorrelationID: parsedLog["correlation_id"].(string),
		Message:       parsedLog["log"].(string),
	}

	logData := LogData{
		Container:    containerID,
		FriendlyName: containerName,
		Timestamp:    timestamp,
		Target:       target,
		Log:          logEntry,
		Error:        logError,
	}

	hash := getLogHash(containerID, logLineStr)
	filter := bson.D{{"_id", hash}}
	update := bson.D{{"$setOnInsert", logData}}

	_, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("Failed to upsert log for container %s: %v", containerID, err)
	}
}
