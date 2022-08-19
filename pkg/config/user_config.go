/**
 * This file is part of Privado OSS.
 *
 * Privado is an open source static code analysis tool to discover data flows in the code.
 * Copyright (C) 2022 Privado, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * For more information, contact support@privado.ai
 */

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Privado-Inc/privado-cli/pkg/auth"
	"github.com/Privado-Inc/privado-cli/pkg/fileutils"
	"github.com/google/uuid"
)

var UserConfig = &UserConfiguration{
	ConfigFile: &UserConfigurationFromFile{
		MetricsEnabled: true,
	},
	SessionId: uuid.NewString(),
}

type UserConfiguration struct {
	ConfigFile       *UserConfigurationFromFile
	UserHash         string
	DockerAccessHash string
	SessionId        string
}

type UserConfigurationFromFile struct {
	MetricsEnabled     bool `json:"metrics"`
	SyncToPrivadoCloud bool `json:"syncToPrivadoCloud"`
}

// Bootstraps user configuration file
// checks for and creates default configuration file if required
func BootstrapUserConfiguration(resetConfig bool) error {
	// check if configuration file exists (skip for reset)
	if !resetConfig {
		if exists, _ := fileutils.DoesFileExists(AppConfig.UserConfigurationFilePath); exists {
			return nil
		}
	}

	// if reset config, update session values that will be saved
	if resetConfig {
		UserConfig.ConfigFile.MetricsEnabled = true
		UserConfig.ConfigFile.SyncToPrivadoCloud = false
	}

	// if not, create directory and file
	if err := os.MkdirAll(AppConfig.ConfigurationDirectory, os.ModePerm); err != nil {
		return err
	}

	if err := SaveUserConfigurationFile(); err != nil {
		return err
	}
	fmt.Println("> Generating configuration file:", AppConfig.UserConfigurationFilePath)

	return nil
}

// loads all required user configuration including from file into UserConfig
func LoadUserConfiguration() {
	// load config from file
	if err := LoadUserConfigurationFile(UserConfig.ConfigFile); err != nil {
		panic(fmt.Sprint(
			fmt.Sprintf("Fatal: cannot load user configuration (%s): %s\n\n", AppConfig.UserConfigurationFilePath, err),
			fmt.Sprintln("To reset privado configuration, simply delete the file and we will generate a new one for you!"),
		))
	}

	// load other configs
	// (move this to another function if these configs increases)
	UserConfig.UserHash = auth.GetUserHash(AppConfig.UserKeyPath)
}

func LoadUserDockerHash(key string) {
	UserConfig.DockerAccessHash = auth.CalculateSHA256Hash(key)
}

// Saves the current UserConfig.ConfigFile to the configuration file
func SaveUserConfigurationFile() error {
	configFileBytes, err := json.MarshalIndent(UserConfig.ConfigFile, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(AppConfig.UserConfigurationFilePath, configFileBytes, 0644); err != nil {
		return err
	}

	return nil
}

func LoadUserConfigurationFile(userConfig *UserConfigurationFromFile) error {
	// read the config file
	data, err := os.ReadFile(AppConfig.UserConfigurationFilePath)
	if err != nil {
		return err
	}

	// load the config file into var
	if err := json.Unmarshal(data, userConfig); err != nil {
		return err
	}

	return nil
}
