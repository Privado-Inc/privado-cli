package main

import (
	"fmt"

	"github.com/Privado-Inc/privado-cli/cmd"
	"github.com/Privado-Inc/privado-cli/pkg/auth"
	"github.com/Privado-Inc/privado-cli/pkg/config"
)

func bootstrap() {
	// bootstrap the userkey UUID
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
