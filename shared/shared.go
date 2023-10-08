package shared

import (
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
