package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/assimoes/rtd-sandbox/logger"
	"github.com/assimoes/rtd-sandbox/shared"
	"github.com/segmentio/kafka-go"
)

var (
	broker       = shared.GetEnv("KAFKA_BROKER", "localhost:9092")
	friendlyName = shared.GetEnv("FRIENDLY_NAME", "monitor")
)

func main() {

	controlCh, controlErrCh := readTopic("control")
	commitCh, commitErrCh := readTopic("commit")

	customLogger := logger.New(friendlyName)

	go errorLogger("control", controlErrCh, customLogger)
	go errorLogger("commit", commitErrCh, customLogger)

	go processDataRequests(controlCh, customLogger)
	go processCommitRequests(commitCh, customLogger)

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

		var data shared.DataRequest
		json.Unmarshal(ctrl.Value, &data)

		if executionID != "" {
			data.ExecutionID = executionID
		}

		customLogger.Log("Kafka", fmt.Sprintf("received data request %s", data.ExecutionID), nil, data.CorrelationID, data.ExecutionID)

		res, err := http.Get(data.Callback + "?correlation_id=" + data.CorrelationID + "&execution_id=" + data.ExecutionID)

		if err != nil {
			customLogger.Log("Kafka", fmt.Sprintf("error calling back the source system: %v", err), err, data.CorrelationID, data.ExecutionID)
			continue
		}

		customLogger.Log("Kafka", fmt.Sprintf("got http status code from source system: %s", res.Status), nil, data.CorrelationID, data.ExecutionID)
	}
}

func processCommitRequests(commitCh chan kafka.Message, customLogger *logger.CustomLogger) {

	for cmt := range commitCh {

		var executionID string

		for _, header := range cmt.Headers {
			if header.Key == "execution_id" {
				executionID = string(header.Value)
				break
			}
		}

		var data shared.CommitRequest
		json.Unmarshal(cmt.Value, &data)

		if executionID != "" {
			data.ExecutionID = executionID
		}

		if data.Commit {
			evt := shared.Event{
				CorrelationID: data.CorrelationID,
				ExecutionID:   data.ExecutionID,
				ServiceName:   data.OriginService,
			}

			customLogger.Log("Kafka", fmt.Sprintf("received event %s", evt.CorrelationID), nil, evt.CorrelationID, evt.ExecutionID)

			evtData, _ := json.Marshal(evt)

			err := publish("e_topic", []kafka.Message{{
				Key:       []byte(evt.CorrelationID),
				Value:     evtData,
				Partition: 0,
			}}, customLogger)

			if err != nil {
				customLogger.Log("Kafka", fmt.Sprintf("error publishing event to event topic: %v", err), err, evt.CorrelationID, evt.ExecutionID)
				continue
			}

			customLogger.Log("Kafka", fmt.Sprintf("published event %s to event topic", evt.CorrelationID), nil, evt.CorrelationID, evt.ExecutionID)
		}
	}
}

func publish(topic string, data []kafka.Message, customLogger *logger.CustomLogger) error {
	k := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Dialer:  kafka.DefaultDialer,
		Topic:   topic,
	})

	if err := k.WriteMessages(context.Background(), data...); err != nil {
		customLogger.Log("Kafka", fmt.Sprintf("Error when publishing to %s topic: %v", topic, err), err, "", "")
		return err
	}

	return nil
}
