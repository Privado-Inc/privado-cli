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
	"runtime"
	"time"

	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of Privado CLI",
	Long:  "Print the current version of Privado CLI",
	Args:  cobra.ExactArgs(0),
	Run:   version,
}

func version(cmd *cobra.Command, args []string) {
	printVersion := Version
	if Version == "dev" {
		printVersion = "Nightly"
	}
	fmt.Printf("Privado CLI: Version %s (%s-%s) \n", printVersion, runtime.GOOS, runtime.GOARCH)

	// Additional info for exclusively this cmd (so 'version' can be called just to print version)
	if cmd.Name() == "version" {
		hasUpdate, updateMessage, err := checkForUpdate()
		if err == nil && hasUpdate {
			fmt.Println()
			fmt.Println(updateMessage)
			time.Sleep(config.AppConfig.SlowdownTime)
			fmt.Println("To use the latest version of Privado CLI, run `privado update`")
			fmt.Println()
		}

		time.Sleep(config.AppConfig.SlowdownTime)
		fmt.Println("For more information, visit", config.AppConfig.PrivadoRepository)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
