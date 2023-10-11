package shared

import (
	"os"
	"time"
)

type CommitRequest struct {
	CorrelationID string `json:"correlation_id"`
	ExecutionID   string `json:"execution_id"`
	OriginService string `json:"origin_service"`
	Commit        bool   `json:"commit"`
}

type DataRequest struct {
	UserID        string    `json:"user_id"`
	Timestamp     time.Time `json:"timestamp"`
	ServiceName   string    `json:"service_name"`
	Callback      string    `json:"callback"`
	CorrelationID string    `json:"correlation_id"`
	ExecutionID   string    `json:"execution_id"`
}

type EventRequest struct {
	ServiceName   string `json:"service_name"`
	CorrelationID string `json:"correlation_id"`
	ExecutionID   string `json:"execution_id"`
}

type DataResponse struct {
	Status        string `json:"status"`
	CorrelationID string `json:"correlation_id"`
}

type Event struct {
	CorrelationID string `json:"correlation_id"`
	ExecutionID   string `json:"execution_id"`
	ServiceName   string `json:"service_name"`
}

type Message struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
}

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

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
