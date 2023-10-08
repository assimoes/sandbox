package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/assimoes/rtd-sandbox/logger" // Import the logger package
	"github.com/assimoes/rtd-sandbox/shared"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var (
	broker       = getEnv("KAFKA_BROKER", "localhost:9092")
	friendlyName = os.Getenv("FRIENDLY_NAME") // Get the friendly name from environment variable
)

func main() {
	// Initialize the custom logger with the friendly name.
	customLogger := logger.New(friendlyName)

	http.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		correlationID := uuid.New().String()

		var dataReq shared.DataRequest

		if err := json.NewDecoder(r.Body).Decode(&dataReq); err != nil {
			customLogger.Log("Kafka", fmt.Sprintf("Error decoding data request: %v", err), err, correlationID, dataReq.ExecutionID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dataReq.CorrelationID = correlationID

		customLogger.Log("Kafka", fmt.Sprintf("Publishing request with correlation ID: %s", correlationID), nil, correlationID, dataReq.ExecutionID)

		w.WriteHeader(http.StatusOK)

		dataRes := shared.DataResponse{
			Status:        "OK",
			CorrelationID: correlationID,
		}

		topicData, _ := json.Marshal(dataReq)

		if err := publish("control", []kafka.Message{
			{
				Partition: 0,
				Key:       []byte(correlationID),
				Value:     topicData,
				Headers: []kafka.Header{
					{Key: "execution_id", Value: []byte(dataReq.ExecutionID)},
				},
			},
		}, customLogger); err != nil {
			customLogger.Log("Kafka", fmt.Sprintf("Error when publishing to control topic: %v", err), err, correlationID, dataReq.ExecutionID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(dataRes)
	})

	http.HandleFunc("/commit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		correlationID := uuid.New().String()

		var commitReq shared.CommitRequest

		if err := json.NewDecoder(r.Body).Decode(&commitReq); err != nil {
			customLogger.Log("Kafka", fmt.Sprintf("Error decoding commit request: %v", err), err, correlationID, commitReq.ExecutionID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var topic string
		if commitReq.Commit {
			topic = "commit"
			customLogger.Log("Kafka", fmt.Sprintf("Publishing commit with correlation ID: %s", correlationID), nil, correlationID, commitReq.ExecutionID)
		} else {
			topic = "cancel"
			customLogger.Log("Kafka", fmt.Sprintf("Publishing cancel with correlation ID: %s", correlationID), nil, correlationID, commitReq.ExecutionID)
		}

		w.WriteHeader(http.StatusOK)

		topicData, _ := json.Marshal(commitReq)

		publish(topic, []kafka.Message{
			{
				Partition: 0,
				Key:       []byte(correlationID),
				Value:     topicData,
			},
		}, customLogger)
	})

	serverAddress := ":3000"
	fmt.Printf("Starting forwarder on %s...\n", serverAddress)

	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		customLogger.Log("Kafka", fmt.Sprintf("Error starting forwarder: %v", err), err, "error-correlation-id", "error-execution-id")
	}
}

func publish(topic string, messages []kafka.Message, customLogger *logger.CustomLogger) error {
	k := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Dialer:  kafka.DefaultDialer,
		Topic:   topic,
	})

	if err := k.WriteMessages(context.Background(), messages...); err != nil {
		customLogger.Log("Kafka", fmt.Sprintf("Error when publishing to %s topic: %v", topic, err), err, "", "")
		return err
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
