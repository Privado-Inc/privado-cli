package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Note: current implementation is based on creating a telemetry instance
// then update using instance methods, however, that will require
// maintaining the instance elsewhere, which can be solved by
// maintaining a singular telemetry instance in the package itself
// similar to config

// creating a DefaultInstance for global updates
var DefaultInstance = InitiateTelemetryInstance()

type Telemetry struct {
	metricMap   map[string]interface{}
	requestBody telemetryRequestBody
}

type telemetryRequestBody struct {
	EventType    string `json:"event_type"`
	EventMessage string `json:"event_message"`
	UserHash     string `json:"user_hash"`
	SessionId    string `json:"session_id"`
}

type telemetryRequestConfig struct {
	url, userHash, sessionId, authenticationKeyHash string
}

func isSupportedMetric(key string) bool {
	switch key {
	case
		"os",
		"arch",
		"cmd",
		"version",
		"warning",
		"error":
		return true
	}

	return false
}

func InitiateTelemetryInstance() *Telemetry {
	var newTelemetryInstance = &Telemetry{
		requestBody: telemetryRequestBody{
			EventType: "PRIVADO_CLI",
		},
	}

	return newTelemetryInstance
}

func (t *Telemetry) RecordAtomicMetric(key string, value interface{}) {
	if isSupportedMetric(key) {
		t.metricMap[key] = fmt.Sprintf("%v", value)
	}
}

func (t *Telemetry) RecordArrayMetric(key string, value interface{}) {
	// perform only if supported metric
	if isSupportedMetric(key) {
		// if key exists, append value, else define value
		if val, ok := t.metricMap[key]; ok {
			// if already an array, append value, else transform value into an array with both values
			if typedVal, isArray := val.([]string); isArray {
				t.metricMap[key] = append(typedVal, fmt.Sprintf("%v", value))
			} else {
				t.metricMap[key] = []string{fmt.Sprintf("%v", val), fmt.Sprintf("%v", value)}
			}
		} else {
			t.metricMap[key] = []string{fmt.Sprintf("%v", value)}
		}
	}
}

func (t *Telemetry) PostRecordedTelemetry(reqConfig telemetryRequestConfig) error {
	t.requestBody.UserHash = reqConfig.userHash
	t.requestBody.SessionId = reqConfig.sessionId

	if metricsJson, err := json.Marshal(t.metricMap); err != nil {
		return err
	} else {
		t.requestBody.EventMessage = string(metricsJson)
	}

	requestBody, err := json.Marshal(t.requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", reqConfig.url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authentication", reqConfig.authenticationKeyHash)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("received non-ok status from telemetry: %d", res.StatusCode)
	}

	return nil
}
