package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/assimoes/rtd-sandbox/shared"
)

type CustomLogger struct {
	friendlyName string
}

func New(friendlyName string) *CustomLogger {
	return &CustomLogger{
		friendlyName: friendlyName,
	}
}

func (c *CustomLogger) Log(target string, message string, err interface{}, correlationID string, executionID string) {

	data := shared.LogData{
		FriendlyName:  c.friendlyName,
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		Target:        target,
		CorrelationID: correlationID,
		ExecutionID:   executionID,
		Log: shared.LogEntry{
			ExecutionID: executionID,
			Message:     message,
		},
		Error: err,
	}

	str, _ := json.Marshal(data)
	log.Println(string(str))
	fmt.Println(string(str))
}
