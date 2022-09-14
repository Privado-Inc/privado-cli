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

package main

import (
	"fmt"

	"github.com/Privado-Inc/privado-cli/cmd"
	"github.com/Privado-Inc/privado-cli/pkg/auth"
	"github.com/Privado-Inc/privado-cli/pkg/ci"
	"github.com/Privado-Inc/privado-cli/pkg/config"
)

func bootstrap() {
	// bootstrap to populate ci session details from env in the ci package
	ci.Bootstrap(config.AppConfig.CIUserIdentifierEnvKey)

	// bootstrap the userkey UUID
	// Any existing "user.key" will override the identified CIUserIdentifier in the previous step
	// Existing key takes precendence. This is intentional as CI users also may want to bootstrap
	// their environment with an existing key, in which case also the set env var is ignored
	err := auth.BootstrapUserKey(config.AppConfig.UserKeyPath, config.AppConfig.UserKeyDirectory)
	if err != nil {
		panic(fmt.Sprintf("Fatal: cannot bootstrap user key: %s", err))
	}

	// bootstrap the configuration file
	err = config.BootstrapUserConfiguration(false)
	if err != nil {
		panic(fmt.Sprintf("Fatal: cannot bootstrap user configuration: %s", err))
	}

	config.LoadUserConfiguration()
}

func main() {
	bootstrap()
	cmd.Execute()
}
