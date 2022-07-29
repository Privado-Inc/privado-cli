package config

var UserConfig *UserConfiguration

type UserConfiguration struct {
	ConfigFile UserConfigurationFromFile
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


	// UserConfig = &UserConfiguration{
	// 	MetricsEnabled:  false,
	// 	SyncToCloudView: false,
	// 	UserHash:        "",
	// }

}
