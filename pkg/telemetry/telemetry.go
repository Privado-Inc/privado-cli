package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Privado-Inc/privado-cli/pkg/config"
)

type Metrics struct {
	metricMap     map[string]interface{}
	requestObject metricRequestObject
}

type metricRequestObject struct {
	EventType    string `json:"event_type"`
	EventMessage string `json:"event_message"`
	UserHash     string `json:"user_hash"`
	SessionId    string `json:"session_id"`
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

func InitiateMetricsInstance() (*Metrics, error) {
	var newMetricInstance = &Metrics{
		requestObject: metricRequestObject{
			EventType: "PRIVADO_CLI",
		},
	}

	return newMetricInstance, nil
}

func (m *Metrics) RecordAtomicMetric(key string, value interface{}) {
	if isSupportedMetric(key) {
		m.metricMap[key] = fmt.Sprintf("%v", value)
	}
}

func (m *Metrics) RecordArrayMetric(key string, value interface{}) {
	// perform only if supported metric
	if isSupportedMetric(key) {
		// if key exists, append value, else define value
		if val, ok := m.metricMap[key]; ok {
			// if already an array, append value, else transform value into an array with both values
			if typedVal, isArray := val.([]string); isArray {
				m.metricMap[key] = append(typedVal, fmt.Sprintf("%v", value))
			} else {
				m.metricMap[key] = []string{fmt.Sprintf("%v", val), fmt.Sprintf("%v", value)}
			}
		} else {
			m.metricMap[key] = []string{fmt.Sprintf("%v", value)}
		}
	}
}

func (m *Metrics) PostRecordedTelemetry(userHash, sessionId string) error {
	m.requestObject.UserHash = userHash
	m.requestObject.SessionId = sessionId

	if metricsJson, err := json.Marshal(m.metricMap); err != nil {
		return err
	} else {
		m.requestObject.EventMessage = string(metricsJson)
	}

	requestBody, err := json.Marshal(m.requestObject)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.AppConfig.PrivadoTelemetryEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authentication", config.UserConfig.DockerAccessHash)

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
