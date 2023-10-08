package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type LogEntry struct {
	ExecutionID string `json:"execution_id"`
	Message     string `json:"message"`
}

// LogData contains the log data to be stored.
type LogData struct {
	Container     string      `json:"container"`
	CorrelationID string      `json:"correlation_id"`
	ExecutionID   string      `json:"execution_id"`
	FriendlyName  string      `json:"friendly_name"`
	Timestamp     string      `json:"timestamp"`
	Target        string      `json:"target"`
	Log           LogEntry    `json:"log"`
	Error         interface{} `json:"error,omitempty"`
}

type CustomLogger struct {
	friendlyName string
}

func New(friendlyName string) *CustomLogger {
	return &CustomLogger{
		friendlyName: friendlyName,
	}
}

func (c *CustomLogger) Log(target string, message string, err interface{}, correlationID string, executionID string) {

	data := LogData{
		FriendlyName:  c.friendlyName,
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		Target:        target,
		CorrelationID: correlationID,
		ExecutionID:   executionID,
		Log: LogEntry{
			ExecutionID: executionID,
			Message:     message,
		},
		Error: err,
	}

	str, _ := json.Marshal(data)
	log.Println(string(str))
	fmt.Println(string(str))
}
