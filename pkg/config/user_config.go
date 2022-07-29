package config

import (
	"fmt"

	"github.com/Privado-Inc/privado/pkg/auth"
)

var UserConfig *UserConfiguration

type UserConfiguration struct {
	ConfigFile *UserConfigurationFromFile
	UserHash   string
}

type UserConfigurationFromFile struct {
	MetricsEnabled  bool `json:"metrics"`
	SyncToCloudView bool `json:"syncToCloudView"`
}

// init function for UserConfig
func init() {
	// check for .privado
	// check for config
	// generate user
	err := auth.BootstrapUserKey(AppConfig.UserKeyPath, AppConfig.UserKeyDirectory)
	if err != nil {
		panic(fmt.Sprintf("Fatal: cannot bootstrap user key: %s", err))
	}

	UserConfig = &UserConfiguration{
		ConfigFile: &UserConfigurationFromFile{},
		UserHash:   auth.GetUserHash(AppConfig.UserKeyPath),
	}
}
