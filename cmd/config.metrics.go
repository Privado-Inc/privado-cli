package cmd

import (
	"fmt"
	"strings"

	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "List, enable, or disable telemetry for Privado CLI",
	Run:   configMetrics,
}

func configMetrics(cmd *cobra.Command, args []string) {
	enableFlag, _ := cmd.Flags().GetBool("enable")
	disableFlag, _ := cmd.Flags().GetBool("disable")

	// if both flags are specified, exit with error
	if enableFlag && disableFlag {
		exit("invalid input: please specify 1 flag", true)
	}

	metricsEnabledTextMap := (map[bool]string{true: "enabled", false: "disabled"})

	// if no flags are specified, show the current configuration
	if !enableFlag && !disableFlag {
		exit(fmt.Sprint(
			fmt.Sprintf("Telemetry for Privado CLI: %s\n", strings.ToUpper(metricsEnabledTextMap[config.UserConfig.ConfigFile.MetricsEnabled])),
			"You can use `--enable` or `--disable` flag to update telemetry preferences",
		), false)
	}

	// if enable flag, enable (set in file anyway if enabled since user is explicitly commanding a write)
	if enableFlag {
		config.UserConfig.ConfigFile.MetricsEnabled = true
	} else if disableFlag {
		config.UserConfig.ConfigFile.MetricsEnabled = false
	}

	if err := config.SaveUserConfigurationFile(); err != nil {
		exit(fmt.Sprintf("Cannot save configuration file: %s", err), true)
	}

	exit(fmt.Sprintf("Telemetry for Privado CLI: %s", strings.ToUpper(metricsEnabledTextMap[config.UserConfig.ConfigFile.MetricsEnabled])), false)
}

func init() {
	metricsCmd.Flags().Bool("enable", false, "Enable telemetry events and performance metrics for Privado CLI")
	metricsCmd.Flags().Bool("disable", false, "Disable telemetry events and performance metrics for Privado CLI")
	// [TODO]: Find a way to keep this and privacy.md in sync
	// metricsCmd.Flags().Bool("list", false, "List down all telemetry events and metrics used by Privado CLI")

	configCmd.AddCommand(metricsCmd)
}
