/**
 * This file is part of Privado OSS.
 *
 * Privado is an open source static code analysis tool to discover data flows in the code.
 * Copyright (C) 2022 Privado, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * For more information, contact support@privado.ai
 */

package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
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
	Recorded    bool
}

type telemetryRequestBody struct {
	EventType    string `json:"event_type"`
	EventMessage string `json:"event_message"`
	UserHash     string `json:"user_hash"`
	SessionId    string `json:"session_id"`
}

type TelemetryRequestConfig struct {
	Url, UserHash, SessionId, AuthenticationKeyHash string
}

func isSupportedMetric(key string) bool {
	switch key {
	case
		"os",
		"arch",
		"cmd",
		"dockerCmd",
		"version",
		"env",
		"ci",
		"ciProvider",
		"didReceiveCloudLinkMessage",
		"didParseCloudLink",
		"didAutoSpawnBrowser",
		"warning",
		"error":
		return true
	}

	return false
}

func InitiateTelemetryInstance() *Telemetry {
	var newTelemetryInstance = &Telemetry{
		metricMap: map[string]interface{}{},
		requestBody: telemetryRequestBody{
			EventType: "PRIVADO_CLI",
		},
	}

	// init with default runtime metrics
	newTelemetryInstance.RecordAtomicMetric("os", runtime.GOOS)
	newTelemetryInstance.RecordAtomicMetric("arch", runtime.GOARCH)

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

func (t *Telemetry) GetRecordedMetrics() map[string]interface{} {
	return t.metricMap
}

func (t *Telemetry) PostRecordedTelemetry(reqConfig TelemetryRequestConfig) error {
	t.requestBody.UserHash = reqConfig.UserHash
	t.requestBody.SessionId = reqConfig.SessionId

	if metricsJson, err := json.MarshalIndent(t.metricMap, "", "    "); err != nil {
		return err
	} else {
		t.requestBody.EventMessage = string(metricsJson)
	}

	requestBody, err := json.Marshal(t.requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", reqConfig.Url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authentication", reqConfig.AuthenticationKeyHash)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("received non-ok status from telemetry: %d", res.StatusCode)
	}

	t.Recorded = true

	return nil
}
