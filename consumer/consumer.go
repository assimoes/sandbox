package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/assimoes/rtd-sandbox/logger"
	"github.com/assimoes/rtd-sandbox/shared"
	"github.com/segmentio/kafka-go"
)

var (
	broker       = shared.GetEnv("KAFKA_BROKER", "localhost:9099")
	friendlyName = shared.GetEnv("FRIENDLY_NAME", "consumer_a")
)

func main() {

	eventCh, eventErrCh := readTopic("e_topic")

	customLogger := logger.New(friendlyName)

	go errorLogger("e_topic", eventErrCh, customLogger)

	go processDataRequests(eventCh, customLogger)

	select {}
}

func readTopic(topic string) (chan kafka.Message, chan error) {
	msgCh, errCh := make(chan kafka.Message, 1000), make(chan error, 1000)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{broker},
		Topic:     topic,
		Partition: 0,
		Dialer:    kafka.DefaultDialer,
	})

	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				errCh <- err
				continue
			}

			msgCh <- msg
		}
	}()

	return msgCh, errCh
}

func errorLogger(topic string, errCh chan error, customLogger *logger.CustomLogger) {
	for err := range errCh {
		customLogger.Log(topic, fmt.Sprintf("error reading from %s topic: %v", topic, err), err, "", "")
	}
}

func processDataRequests(controlCh chan kafka.Message, customLogger *logger.CustomLogger) {
	for ctrl := range controlCh {

		var executionID string

		for _, header := range ctrl.Headers {
			if header.Key == "execution_id" {
				executionID = string(header.Value)
				break
			}
		}

		var data shared.EventRequest
		json.Unmarshal(ctrl.Value, &data)

		if executionID != "" {
			data.ExecutionID = executionID
		}

		customLogger.Log("Kafka", fmt.Sprintf("consumed event request %s", data.ExecutionID), nil, data.CorrelationID, data.ExecutionID)

	}
}
