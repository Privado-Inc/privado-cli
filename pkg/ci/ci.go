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

package ci

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// this directive is necessary to load file in this variable;
// it will also include the file automatically during build
// ref about embed directive: https://pkg.go.dev/embed@master
//go:embed providers.json
var CIProvidersFile []byte

// Session CI Configuration
var CISessionConfig *CISession

// Global CI Configuration
var CIConfig *CI

type CI struct {
	CIIdentifierEnvKeys []string
	Providers           *[]Provider
}

type CISession struct {
	IsCI            bool
	RepositoryValue string
	Provider        *Provider
}

type Provider struct {
	Name           string   `json:"name"`
	Identifier     string   `json:"identifier"`
	RepositoryKeys []string `json:"keys"`
}

// populate values for CIConfig
// it will carry values that the package provides
// like constants, without any other function calls
func init() {
	CIConfig = &CI{
		CIIdentifierEnvKeys: []string{
			"CI",
			"CONTINUOUS_INTEGRATION",
		},
	}

	if err := json.Unmarshal(CIProvidersFile, &CIConfig.Providers); err != nil {
		fmt.Println("> Could not parse CI providers from `providers.json`")
	}
}
