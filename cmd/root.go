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
 *
 */

package cmd

import (
	"fmt"
	"os"
	"strings"

	// homedir "github.com/mitchellh/go-homedir"

	"github.com/Privado-Inc/privado-cli/pkg/ci"
	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/Privado-Inc/privado-cli/pkg/telemetry"
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "privado",
	Short: "Privado is a CLI tool that scans & monitors your repositories to build privacy, transparency reports & finds privacy issues",
	Long:  "Privado is a CLI tool that scans & monitors your repositories to build privacy, transparency reports & finds privacy issues. \nFind more at: https://github.com/Privado-Inc/privado",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exit(fmt.Sprintln(err), true)
	}

	defer func() {
		// if panic occurred
		if err := recover(); err != nil {
			// only if we have a docker access hash
			if config.UserConfig.DockerAccessHash != "" {
				// if defaultInstance is already sent, create another, else append to error and send
				if !telemetry.DefaultInstance.Recorded {
					telemetry.DefaultInstance.RecordArrayMetric("error", err)
					telemetryPostRun(nil)
				} else {
					t := telemetry.InitiateTelemetryInstance()
					t.RecordAtomicMetric("version", Version)
					t.RecordArrayMetric("error", err)
					telemetryPostRun(t)
				}
			}
		}
	}()
}

func telemetryPreRun(t *telemetry.Telemetry) {
	if t == nil {
		t = telemetry.DefaultInstance
	}

	t.RecordAtomicMetric("version", Version)
	t.RecordAtomicMetric("cmd", strings.Join(os.Args, " "))
	t.RecordAtomicMetric("ci", ci.CISessionConfig.IsCI)
	if ci.CISessionConfig.IsCI && ci.CISessionConfig.Provider != nil {
		t.RecordAtomicMetric("ciProvider", ci.CISessionConfig.Provider.Name)
	}
}

func telemetryPostRun(t *telemetry.Telemetry) {
	if t == nil {
		t = telemetry.DefaultInstance
	}

	t.PostRecordedTelemetry(telemetry.TelemetryRequestConfig{
		Url:                   config.AppConfig.PrivadoTelemetryEndpoint,
		UserHash:              config.UserConfig.UserHash,
		SessionId:             config.UserConfig.SessionId,
		AuthenticationKeyHash: config.UserConfig.DockerAccessHash,
	})
}

func exit(msg string, error bool) {
	fmt.Println(msg)
	if error {
		telemetry.DefaultInstance.RecordArrayMetric("error", msg)
	}

	if !telemetry.DefaultInstance.Recorded && config.UserConfig.DockerAccessHash != "" {
		telemetryPostRun(nil)
	}

	if error {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
