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
	metricsCmd.MarkFlagsMutuallyExclusive("enable", "disable")
	// [TODO]: Find a way to keep this and privacy.md in sync
	// metricsCmd.Flags().Bool("list", false, "List down all telemetry events and metrics used by Privado CLI")

	configCmd.AddCommand(metricsCmd)
}
