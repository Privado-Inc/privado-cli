package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/Privado-Inc/privado-cli/pkg/docker"
	"github.com/Privado-Inc/privado-cli/pkg/fileutils"
	"github.com/Privado-Inc/privado-cli/pkg/telemetry"
	"github.com/Privado-Inc/privado-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <repository>",
	Short: "Scan a codebase or repository to identify privacy issues and generate compliance reports",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		telemetry.DefaultInstance.RecordAtomicMetric("version", Version)
		telemetry.DefaultInstance.RecordAtomicMetric("cmd", strings.Join(os.Args, " "))
	},
	Run: scan,
	PostRun: func(cmd *cobra.Command, args []string) {
		telemetryPostRun(nil)
	},
}

func defineScanFlags(cmd *cobra.Command) {
	scanCmd.Flags().StringP("rules", "r", "", "Specifies the rule directory to be passed to privado-core for scanning. These external rules are merged with the default set of rules that Privado defines")
	scanCmd.Flags().BoolP("ignore-default-rules", "i", false, "If specified, the default rules are ignored and only the specified rules (-r) are considered")
	scanCmd.Flags().Bool("skip-dependency-download", false, "When specified, the engine skips downloading all locally unavailable dependencies. Skipping dependency download can yield incomplete results")
	scanCmd.Flags().Bool("overwrite", false, "If specified, the warning prompt for existing scan results is disabled and any existing results are overwritten")
	scanCmd.Flags().Bool("debug", false, "Enables privado-core image output in debug mode")
}

func scan(cmd *cobra.Command, args []string) {
	repository := args[0]
	debug, _ := cmd.Flags().GetBool("debug")
	overwriteResults, _ := cmd.Flags().GetBool("overwrite")
	skipDependencyDownload, _ := cmd.Flags().GetBool("skip-dependency-download")

	externalRules, _ := cmd.Flags().GetString("rules")
	if externalRules != "" {
		externalRules = fileutils.GetAbsolutePath(externalRules)
		externalRulesExists, _ := fileutils.DoesFileExists(externalRules)
		if !externalRulesExists {
			exit(fmt.Sprintf("Could not validate the rules directory: %s", externalRules), true)
		}
	}

	ignoreDefaultRules, _ := cmd.Flags().GetBool("ignore-default-rules")
	if ignoreDefaultRules && externalRules == "" {
		exit(fmt.Sprint(
			"Default rules cannot be ignored without any external rules.\n",
			"You can specify your own rules using the `-r` option.\n\n",
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
			fmt.Println("\n> Rescan will overwrite existing results and progress")
			confirm, _ := utils.ShowConfirmationPrompt("Continue?")
			if !confirm {
				exit("Terminating..", false)
			}
			fmt.Println()
		}
	}

	fmt.Println("> Scanning directory:", fileutils.GetAbsolutePath(repository))

	if dockerAccessKey, err := docker.GetPrivadoDockerAccessKey(true); err != nil || dockerAccessKey == "" {
		exit(fmt.Sprintf("Cannot fetch docker access key: %v \nPlease try again or raise an issue at %s", err, config.AppConfig.PrivadoRepository), true)
	} else {
		config.LoadUserDockerHash(dockerAccessKey)
	}

	// "always pass -ir: even when internal rules are ignored (-i)"
	commandArgs := []string{config.AppConfig.Container.SourceCodeVolumeDir, "-ir", config.AppConfig.Container.InternalRulesVolumeDir}

	// run image with options
	err = docker.RunImage(
		docker.OptionWithLatestImage(false), // because we already pull the image for access-key (with pullImage parameter)
		docker.OptionWithArgs(commandArgs),
		docker.OptionWithAttachedOutput(),
		docker.OptionWithSourceVolume(fileutils.GetAbsolutePath(repository)),
		docker.OptionWithUserConfigVolume(config.AppConfig.UserConfigurationFilePath),
		docker.OptionWithUserKeyVolume(config.AppConfig.UserKeyPath),
		docker.OptionWithIgnoreDefaultRules(ignoreDefaultRules),
		docker.OptionWithExternalRulesVolume(externalRules),
		docker.OptionWithSkipDependencyDownload(skipDependencyDownload),
		docker.OptionWithPackageCacheVolume(config.AppConfig.M2PackageCacheDirectory),
		docker.OptionWithEnvironmentVariables([]docker.EnvVar{
			{Key: "PRIVADO_VERSION_CLI", Value: Version},
			{Key: "PRIVADO_HOST_SCAN_DIR", Value: fileutils.GetAbsolutePath(repository)},
			{Key: "PRIVADO_USER_HASH", Value: config.UserConfig.UserHash},
			{Key: "PRIVADO_SESSION_ID", Value: config.UserConfig.SessionId},
			{Key: "PRIVADO_SYNC_TO_CLOUD", Value: strings.ToUpper(strconv.FormatBool(config.UserConfig.ConfigFile.SyncToPrivadoCloud))},
			{Key: "PRIVADO_METRICS_ENABLED", Value: strings.ToUpper(strconv.FormatBool(config.UserConfig.ConfigFile.MetricsEnabled))},
		}),
		docker.OptionWithDebug(debug),
	)
	if err != nil {
		exit(fmt.Sprintf("Received error: %s", err), true)
	}
}

func init() {
	defineScanFlags(scanCmd)
	rootCmd.AddCommand(scanCmd)
}
