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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Privado-Inc/privado-cli/pkg/ci"
	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/Privado-Inc/privado-cli/pkg/docker"
	"github.com/Privado-Inc/privado-cli/pkg/fileutils"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <repository>",
	Short: "Sync scan results with Privado Dashboard",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		telemetryPreRun(nil)
	},
	Run: upload,
	PostRun: func(cmd *cobra.Command, args []string) {
		telemetryPostRun(nil)
	},
}

func upload(cmd *cobra.Command, args []string) {
	repository := args[0]
	debug, _ := cmd.Flags().GetBool("debug")

	hasUpdate, updateMessage, err := checkForUpdate()
	if err == nil && hasUpdate {
		fmt.Println(updateMessage)
		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println("To use the latest version of Privado CLI, run `privado update`")
		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println()
	}

	fmt.Println("> Uploading results for directory:", fileutils.GetAbsolutePath(repository))
	time.Sleep(config.AppConfig.SlowdownTime)

	resultsPath := filepath.Join(fileutils.GetAbsolutePath(repository), config.AppConfig.PrivacyResultsPathSuffix)
	if exists, _ := fileutils.DoesFileExists(resultsPath); !exists {
		exit(fmt.Sprint(
			fmt.Sprintf("Cannot find scan results (%s) in the specified directory\n", config.AppConfig.PrivacyResultsPathSuffix),
			"Run 'privado scan <dir>' instead\n\n",
			"Run 'privado help' for more information.",
		), true)
	}

	if dockerAccessKey, err := docker.GetPrivadoDockerAccessKey(true); err != nil || dockerAccessKey == "" {
		exit(fmt.Sprintf("Cannot fetch docker access key: %v \nPlease try again or raise an issue at %s", err, config.AppConfig.PrivadoRepository), true)
	} else {
		config.LoadUserDockerHash(dockerAccessKey)
	}

	command := []string{
		config.AppConfig.Container.PrivadoCoreBinPath,
		"upload",
	}

	commandArgs := []string{config.AppConfig.Container.SourceCodeVolumeDir}

	// run image with options
	err = docker.RunImage(
		docker.OptionWithLatestImage(false), // because we already pull the image for access-key (with pullImage parameter)
		docker.OptionWithEntrypoint(command),
		docker.OptionWithArgs(commandArgs),
		docker.OptionWithAttachedOutput(),
		docker.OptionWithSourceVolume(fileutils.GetAbsolutePath(repository)),
		docker.OptionWithUserKeyVolume(config.AppConfig.UserKeyPath),
		docker.OptionWithDebug(debug),
		docker.OptionWithEnvironmentVariables([]docker.EnvVar{
			{Key: "CI", Value: strings.ToUpper(strconv.FormatBool(ci.CISessionConfig.IsCI))},
			{Key: "PRIVADO_VERSION_CLI", Value: Version},
			{Key: "PRIVADO_HOST_SCAN_DIR", Value: fileutils.GetAbsolutePath(repository)},
			{Key: "PRIVADO_USER_HASH", Value: config.UserConfig.UserHash},
			{Key: "PRIVADO_SESSION_ID", Value: config.UserConfig.SessionId},
			{Key: "PRIVADO_SYNC_TO_CLOUD", Value: strings.ToUpper(strconv.FormatBool(config.UserConfig.ConfigFile.SyncToPrivadoCloud))},
			{Key: "PRIVADO_METRICS_ENABLED", Value: strings.ToUpper(strconv.FormatBool(config.UserConfig.ConfigFile.MetricsEnabled))},
		}),
		docker.OptionWithAutoSpawnBrowserOnURLMessages([]string{
			"> Continue to view results on:",
		}),
		docker.OptionWithInterrupt(),
	)
	if err != nil {
		exit(fmt.Sprintf("Received error: %s", err), true)
	}
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
