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
	"github.com/Privado-Inc/privado-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <repository>",
	Short: "Scan a codebase or repository to identify privacy issues and generate compliance reports",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		telemetryPreRun(nil)
	},
	Run: scan,
	PostRun: func(cmd *cobra.Command, args []string) {
		telemetryPostRun(nil)
	},
}

func defineScanFlags(cmd *cobra.Command) {
	scanCmd.Flags().StringP("config", "c", "", "Specifies the config (with rules) directory to be passed to privado-core for scanning. These external rules and configurations are merged with the default set that Privado defines")
	scanCmd.Flags().BoolP("ignore-default-rules", "i", false, "If specified, the default rules are ignored and only the specified rule configurations (-c) are considered")
	scanCmd.Flags().Bool("skip-dependency-download", false, "When specified, the engine skips downloading all locally unavailable dependencies. Skipping dependency download can yield incomplete results")
	scanCmd.Flags().Bool("disable-deduplication", false, "When specified, the engine does not remove duplicate and subset dataflows. This option is useful if you wish to review all flows (including duplicates) manually")

	scanCmd.Flags().Bool("upload", false, "If specified, will automatically attempt to upload the scan result to Privado Dashboard")
	scanCmd.Flags().Bool("skip-upload", false, "If specified, the result artifacts will not be uploaded to Privado Dashboard")
	scanCmd.MarkFlagsMutuallyExclusive("upload", "skip-upload")

	scanCmd.Flags().Bool("overwrite", false, "If specified, the warning prompt for existing scan results is disabled and any existing results are overwritten")
	scanCmd.Flags().Bool("debug", false, "Enables privado-core image output in debug mode")
	scanCmd.Flags().String("jvm-args", "", "Specifies the JVM arguments to be passed to the scan engine; sets the 'JAVA_TOOL_OPTIONS' environment variable")
	scanCmd.Flags().Bool("enable-experiments", false, "Flag to enable experimental features")
	scanCmd.Flags().Bool("enable-javascript", false, "Experimental: When specified, enables the beta code scanner for javascript. Use with '--enable-experiments'")
	scanCmd.Flags().Bool("disable-runtime-semantics", false, "Experimental: If specified, the semantics engine won't generate semantic at runtime")
	scanCmd.Flags().Bool("disable-this-filtering", false, "Experimental: If specified, filtering of flow using 'this filtering algorithm' will be avoided")
	scanCmd.Flags().Bool("disable-flow-separation-by-data-element", false, "Experimental: If specified, filtering of flow using 'flow separation by data element algorithm' will be avoided")
	scanCmd.Flags().Bool("disable-2nd-level-closure", false, "Experimental: If specified, 2nd level source derivation will be turned on")
	scanCmd.Flags().Bool("generate-unresolved-name-report", false, "Flag to enable generation unresolved method name reports")
	scanCmd.Flags().Bool("test-output", false, "Flag to generate unfiltered flow output getting direclty from joern")
}

func scan(cmd *cobra.Command, args []string) {
	repository := args[0]
	debug, _ := cmd.Flags().GetBool("debug")
	overwriteResults, _ := cmd.Flags().GetBool("overwrite")
	skipDependencyDownload, _ := cmd.Flags().GetBool("skip-dependency-download")
	disableDeduplication, _ := cmd.Flags().GetBool("disable-deduplication")
	explicitUpload, _ := cmd.Flags().GetBool("upload")
	explicitSkipUpload, _ := cmd.Flags().GetBool("skip-upload")
	jvmArgs, _ := cmd.Flags().GetString("jvm-args")
	experimentalEnabled, _ := cmd.Flags().GetBool("enable-experiments")
	experimentalJavascriptEnabled, _ := cmd.Flags().GetBool("enable-javascript")
	disableRunTimeSemantics, _ := cmd.Flags().GetBool("disable-runtime-semantics")
	disableThisFiltering, _ := cmd.Flags().GetBool("disable-this-filtering")
	disableFlowSeperationByDataElement, _ := cmd.Flags().GetBool("disable-flow-separation-by-data-element")
	disable2ndLevelClosure, _ := cmd.Flags().GetBool("disable-2nd-level-closure")
	generateUnresolvedNameReport, _ := cmd.Flags().GetBool("generate-unresolved-name-report")
	testOutput, _ := cmd.Flags().GetBool("test-output")

	externalRules, _ := cmd.Flags().GetString("config")
	if externalRules != "" {
		externalRules = fileutils.GetAbsolutePath(externalRules)
		externalRulesExists, _ := fileutils.DoesFileExists(externalRules)
		if !externalRulesExists {
			exit(fmt.Sprintf("Could not validate the config directory: %s", externalRules), true)
		}
	}

	ignoreDefaultRules, _ := cmd.Flags().GetBool("ignore-default-rules")
	if ignoreDefaultRules && externalRules == "" {
		exit(fmt.Sprint(
			"Default rules cannot be ignored without any external config.\n",
			"You can specify your own rules and config using the `-c or --config` option.\n\n",
			"For more info, run: 'privado help'\n",
		), true)
	}

	hasUpdate, updateMessage, err := checkForUpdate()
	if err == nil && hasUpdate {
		fmt.Println(updateMessage)
		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println("To use the latest version of Privado CLI, run `privado update`")
		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println()
	}

	// if overwrite flag is not specified, check for existing results
	if !overwriteResults {
		resultsPath := filepath.Join(fileutils.GetAbsolutePath(repository), config.AppConfig.PrivacyResultsPathSuffix)
		if exists, _ := fileutils.DoesFileExists(resultsPath); exists {
			fmt.Printf("> Scan report already exists (%s)\n", config.AppConfig.PrivacyResultsPathSuffix)
			fmt.Println("\n> Rescan will overwrite existing results")
			confirm, _ := utils.ShowConfirmationPrompt("Continue?")
			if !confirm {
				exit("Terminating..", false)
			}
			fmt.Println()
		}
	}

	if !experimentalEnabled && (experimentalJavascriptEnabled || disableRunTimeSemantics || disableThisFiltering || disableFlowSeperationByDataElement || disable2ndLevelClosure) {
		exit(fmt.Sprint(
			"Experimental features cannot be used without the `--enable-experiments` flag.\n\n",
			"For more info, run: 'privado help'\n",
		), true)
	}

	fmt.Println("> Scanning directory:", fileutils.GetAbsolutePath(repository))

	if dockerAccessKey, err := docker.GetPrivadoDockerAccessKey(true); err != nil || dockerAccessKey == "" {
		exit(fmt.Sprintf("Cannot fetch docker access key: %v \nPlease try again or raise an issue at %s", err, config.AppConfig.PrivadoRepository), true)
	} else {
		config.LoadUserDockerHash(dockerAccessKey)
	}

	// "always pass -ic: even when internal rules are ignored (-i)"
	commandArgs := []string{
		config.AppConfig.Container.SourceCodeVolumeDir,
		"-ic",
		config.AppConfig.Container.InternalRulesVolumeDir,
	}

	if explicitUpload {
		commandArgs = append(commandArgs, "--upload")
	} else if explicitSkipUpload {
		commandArgs = append(commandArgs, "--skip-upload")
	}

	if experimentalJavascriptEnabled {
		commandArgs = append(commandArgs, "--enablejs")
	}

	if disableRunTimeSemantics {
		commandArgs = append(commandArgs, "-drs")
	}

	if disableFlowSeperationByDataElement {
		commandArgs = append(commandArgs, "-dfsde")
	}

	if disableThisFiltering {
		commandArgs = append(commandArgs, "-dtf")
	}

	if disable2ndLevelClosure {
		commandArgs = append(commandArgs, "-d2lc")
	}

	if generateUnresolvedNameReport {
		commandArgs = append(commandArgs, "-ur")
	}

	if testOutput {
		commandArgs = append(commandArgs, "-tout")
	}

	// run image with options
	err = docker.RunImage(
		docker.OptionWithLatestImage(false), // because we already pull the image for access-key (with pullImage parameter)
		docker.OptionWithArgs(commandArgs),
		docker.OptionWithAttachedOutput(),
		docker.OptionWithSourceVolume(fileutils.GetAbsolutePath(repository)),
		docker.OptionWithUserConfigVolume(config.AppConfig.UserConfigurationFilePath),
		docker.OptionWithUserKeyVolume(config.AppConfig.UserKeyPath),
		docker.OptionWithPackageCacheVolumes(),
		docker.OptionWithExternalRulesVolume(externalRules),
		docker.OptionWithIgnoreDefaultRules(ignoreDefaultRules),
		docker.OptionWithSkipDependencyDownload(skipDependencyDownload),
		docker.OptionWithDisabledDeduplication(disableDeduplication),

		docker.OptionWithDebug(debug),
		docker.OptionWithEnvironmentVariables([]docker.EnvVar{
			{Key: "CI", Value: strings.ToUpper(strconv.FormatBool(ci.CISessionConfig.IsCI))},
			{Key: "PRIVADO_VERSION_CLI", Value: Version},
			{Key: "PRIVADO_HOST_SCAN_DIR", Value: fileutils.GetAbsolutePath(repository)},
			{Key: "PRIVADO_USER_HASH", Value: config.UserConfig.UserHash},
			{Key: "PRIVADO_SESSION_ID", Value: config.UserConfig.SessionId},
			{Key: "PRIVADO_SYNC_TO_CLOUD", Value: strings.ToUpper(strconv.FormatBool(config.UserConfig.ConfigFile.SyncToPrivadoCloud))},
			{Key: "PRIVADO_METRICS_ENABLED", Value: strings.ToUpper(strconv.FormatBool(config.UserConfig.ConfigFile.MetricsEnabled))},
			{Key: "JAVA_TOOL_OPTIONS", Value: jvmArgs},
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
	defineScanFlags(scanCmd)
	rootCmd.AddCommand(scanCmd)
}
