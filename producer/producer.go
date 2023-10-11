package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/assimoes/rtd-sandbox/logger" // Import the logger package
	"github.com/assimoes/rtd-sandbox/shared"
	"github.com/google/uuid"
)

var (
	forwarderURL  = shared.GetEnv("FORWARDER_URL", "http://localhost:3000")
	externalName  = shared.GetEnv("EXTERNAL_NAME", "localhost")
	externalPort  = shared.GetEnv("EXTERNAL_PORT", "8888")
	tickerTimeout = 5 * time.Second
	friendlyName  = shared.GetEnv("FRIENDLY_NAME", "producer_a")
	customLogger  *logger.CustomLogger
)

func main() {

	ticker := time.NewTicker(tickerTimeout)
	defer ticker.Stop()

	// Initialize the custom logger with the friendly name.
	customLogger = logger.New(friendlyName)

	go func() {
		for range ticker.C {
			executionID, _ := uuid.NewUUID()
			data := createDataRequest(executionID.String())
			if err := sendData(data, customLogger); err != nil {
				customLogger.Log("forwarder", fmt.Sprintf("Error sending data: %v", err), err, data.CorrelationID, executionID.String())
			}
		}
	}()

	http.HandleFunc("/callback", callback)
	err := http.ListenAndServe(":"+externalPort, nil)
	if err != nil {
		customLogger.Log("forwarder", fmt.Sprintf("Error starting HTTP server: %v", err), err, "error-correlation-id", "error-execution-id")
	}
}

func createDataRequest(executionID string) shared.DataRequest {

	data := shared.DataRequest{
		UserID:      strconv.Itoa(rand.Int()),
		Timestamp:   time.Now(),
		ServiceName: externalName,
		Callback:    fmt.Sprintf("http://%s:%s/callback", externalName, externalPort),
		ExecutionID: executionID,
	}
	return data
}

func callback(w http.ResponseWriter, r *http.Request) {
	correlationID := r.URL.Query().Get("correlation_id")
	executionID := r.URL.Query().Get("execution_id")

	if correlationID == "" {
		customLogger.Log("error", "Missing correlation id", nil, "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if executionID == "" {
		customLogger.Log("error", "Missing execution id", nil, "", "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cr := shared.CommitRequest{
		CorrelationID: correlationID,
		ExecutionID:   executionID,
		Commit:        randBool(),
	}

	if err := commit(cr, customLogger); err != nil {
		customLogger.Log("error", fmt.Sprintf("error committing message to forwarder: %v", err), err, correlationID, executionID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func commit(cr shared.CommitRequest, customLogger *logger.CustomLogger) error {
	data, err := json.Marshal(cr)
	if err != nil {
		return err
	}

	res, err := http.Post(forwarderURL+"/commit", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//_, _ = io.ReadAll(res.Body)
	customLogger.Log("forwarder", fmt.Sprintf("Response code from forwarder: %s", res.Status), nil, cr.CorrelationID, cr.ExecutionID)

	return nil
}

func sendData(data shared.DataRequest, customLogger *logger.CustomLogger) error {
	dataBytes, _ := json.Marshal(data)
	res, err := http.Post(forwarderURL+"/request", "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	customLogger.Log("forwarder", fmt.Sprintf("Response code from forwarder: %s", res.Status), nil, data.CorrelationID, data.ExecutionID)

	return nil
}

func randBool() bool {
	return rand.Intn(2) == 0
}
