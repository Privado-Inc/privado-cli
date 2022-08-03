package cmd

import (
	"fmt"
	"os"

	// homedir "github.com/mitchellh/go-homedir"

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
