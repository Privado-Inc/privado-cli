package docker

import "github.com/Privado-Inc/privado/pkg/config"

type containerVolumes struct {
	userKeyVolumeEnabled, dockerKeyVolumeEnabled, sourceCodeVolumeEnabled, externalRulesVolumeEnabled bool
	userKeyVolumeHost, dockerKeyVolumeHost, sourceCodeVolumeHost, externalRulesVolumeHost             string
}

// type containerPorts struct {
// 	webPortEnabled bool
// 	webPortHost    int
// }

type RunImageOption func(opts *runImageHandler)

type runImageHandler struct {
	args    []string
	volumes containerVolumes
	// ports             containerPorts
	setupInterrupt    bool
	spawnWebBrowser   bool
	attachOutput      bool
	exitErrorMessages []string
}

func newRunImageHandler(opts []RunImageOption) runImageHandler {
	// defaults here
	rh := runImageHandler{}
	for _, opt := range opts {
		opt(&rh)
	}
	return rh
}

// Prepend option functions with "Option"

func OptionWithArgs(args []string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.args = args
	}
}

func OptionWithUserKeyVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.userKeyVolumeEnabled = true
		rh.volumes.userKeyVolumeHost = volumeHost
	}
}

func OptionWithDockerKeyVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.dockerKeyVolumeEnabled = true
		rh.volumes.dockerKeyVolumeHost = volumeHost
	}
}

func OptionWithSourceVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.sourceCodeVolumeEnabled = true
		rh.volumes.sourceCodeVolumeHost = volumeHost
	}
}

func OptionWithExternalRulesVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		if volumeHost != "" {
			rh.volumes.externalRulesVolumeEnabled = true
			rh.volumes.externalRulesVolumeHost = volumeHost
			rh.args = append(rh.args, "-er", config.AppConfig.Container.ExternalRulesVolumeDir)
		}
	}
}

func OptionWithDefaultRules(useDefaultRules bool) RunImageOption {
	return func(rh *runImageHandler) {
		if useDefaultRules {
			rh.args = append(rh.args, "-ir", config.AppConfig.Container.InternalRulesVolumeDir)
		}
	}
}

func OptionWithInterrupt() RunImageOption {
	return func(rh *runImageHandler) {
		rh.setupInterrupt = true
	}
}

func OptionWithAutoSpawnBrowser() RunImageOption {
	return func(rh *runImageHandler) {
		rh.spawnWebBrowser = true
	}
}

func OptionWithExitErrorMessages(messages []string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.exitErrorMessages = messages
	}
}

func OptionWithAttachedOutput() RunImageOption {
	return func(rh *runImageHandler) {
		rh.attachOutput = true
	}
}

func OptionWithDebug(isDebug bool) RunImageOption {
	return func(rh *runImageHandler) {
		// currently only enable output in debug mode
		if isDebug {
			rh.attachOutput = true
			rh.args = append(rh.args, "--debug")
		}
	}
}
