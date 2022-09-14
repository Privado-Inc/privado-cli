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
	"os"
	"strconv"
	"strings"
)

// this directive is necessary to load file in this variable;
// it will also include the file automatically during build
// ref about embed directive: https://pkg.go.dev/embed@master
//go:embed providers.json
var CIProvidersFile []byte

// Session CI Configuration
var CISessionConfig *CISession = &CISession{}

// Global CI Configuration
var CIConfig *CI = &CI{}

type CI struct {
	CIIdentifierEnvKeys     []string
	Providers               *[]Provider
	CustomUserIdentifierKey string
}

type CISession struct {
	IsCI           bool
	UserIdentifier string
	Provider       *Provider
}

type Provider struct {
	// name of the ci provider
	Name string `json:"name"`

	// defines the distinct env key that can be used to
	// identify the CI provider in the ci environment
	Identifier string `json:"identifier"`

	// defines the env keys that can be used to
	// identify the user in the ci environment
	UserKeys []string `json:"keys"`
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

// populates session information in CISessionConfig
// Session values do not populate automatically with
// an intent of required custom loaders that users
// might want to load something before bootstrapping
func Bootstrap(customUserIdentifierKey string) {
	CIConfig.CustomUserIdentifierKey = customUserIdentifierKey

	// detect ci env
	CISessionConfig.IsCI = IsCIEnvironment()
	if CISessionConfig.IsCI {
		fmt.Println("> Detected CI environment")

		// detect provider
		CISessionConfig.Provider = IdentifyCIProvider()
		if CISessionConfig.Provider != nil {
			fmt.Println("> Identified CI provider:", CISessionConfig.Provider.Name)
		}

		// if custom user identifier is defined - use that to attempt to get value
		// else if provider is identified, use that to get value from ci env
		if customUserId := os.Getenv(CIConfig.CustomUserIdentifierKey); customUserId != "" {
			fmt.Println("> Detected set", CIConfig.CustomUserIdentifierKey)
			CISessionConfig.UserIdentifier = customUserId
		} else if CISessionConfig.Provider != nil {
			CISessionConfig.UserIdentifier = CISessionConfig.Provider.GetUserIdentifierFromCIEnvironment()
		}

		if CISessionConfig.UserIdentifier != "" {
			fmt.Println("> Identified CI user:", CISessionConfig.UserIdentifier)
		}
	}
}

func IsCIEnvironment() bool {
	for _, key := range CIConfig.CIIdentifierEnvKeys {
		isCI, _ := strconv.ParseBool(os.Getenv(key))
		if isCI {
			return true
		}
	}
	return false
}

func IdentifyCIProvider() *Provider {
	if err := json.Unmarshal(CIProvidersFile, &CIConfig.Providers); err != nil {
		fmt.Println("Could not identify the CI provider")
	}

	for _, provider := range *CIConfig.Providers {
		if _, exists := os.LookupEnv(provider.Identifier); exists {
			return &provider
		}
	}

	return nil
}

func (provider *Provider) GetUserIdentifierFromCIEnvironment() string {
	values := []string{}

	for _, key := range provider.UserKeys {
		val := os.Getenv(key)
		if val != "" {
			values = append(values, val)
		}
	}

	if len(values) == 0 {
		return ""
	}

	return strings.Join(values, "-")
}
