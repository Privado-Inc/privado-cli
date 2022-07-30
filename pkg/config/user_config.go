package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Privado-Inc/privado/pkg/auth"
	"github.com/Privado-Inc/privado/pkg/fileutils"
)

var UserConfig = &UserConfiguration{
	ConfigFile: &UserConfigurationFromFile{
		MetricsEnabled: true,
	},
}

type UserConfiguration struct {
	ConfigFile *UserConfigurationFromFile
	UserHash   string
}

type UserConfigurationFromFile struct {
	MetricsEnabled     bool `json:"metrics"`
	SyncToPrivadoCloud bool `json:"syncToPrivadoCloud"`
}

// Bootstraps user configuration file
// checks for and creates default configuration file if required
func BootstrapUserConfiguration(resetConfig bool) error {
	// check if configuration file exists
	if !resetConfig {
		if exists, _ := fileutils.DoesFileExists(AppConfig.UserConfigurationFilePath); exists {
			return nil
		}
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
			fmt.Sprintln("To reset privado configuration, run `privado --reset-config`"),
		))
	}

	// load other configs
	// (move this to another function if these configs increases)
	UserConfig.UserHash = auth.GetUserHash(AppConfig.UserKeyPath)
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
