package docker

import "github.com/Privado-Inc/privado/pkg/config"

type containerVolumes struct {
	licenseVolumeEnabled, sourceCodeVolumeEnabled, externalRulesVolumeEnabled bool
	licenseVolumeHost, sourceCodeVolumeHost, externalRulesVolumeHost          string
}

type containerPorts struct {
	webPortEnabled bool
	webPortHost    int
}

type RunImageOption func(opts *runImageHandler)

type runImageHandler struct {
	args               []string
	volumes            containerVolumes
	ports              containerPorts
	setupInterrupt     bool
	spawnWebBrowser    bool
	progressLoader     bool
	duringLoadMessages []string
	afterLoadMessages  []string
	attachOutput       bool
	exitErrorMessages  []string
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

func OptionWithLicenseVolume(volumeHost string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.volumes.licenseVolumeEnabled = true
		rh.volumes.licenseVolumeHost = volumeHost
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

func OptionWithWebPort(portHost int) RunImageOption {
	return func(rh *runImageHandler) {
		rh.ports.webPortEnabled = true
		rh.ports.webPortHost = portHost
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

func OptionWithProgressLoader(duringLoadMessages []string, afterLoadMessages []string) RunImageOption {
	return func(rh *runImageHandler) {
		rh.progressLoader = true
		if len(duringLoadMessages) > 0 {
			rh.duringLoadMessages = duringLoadMessages
		}
		if len(afterLoadMessages) > 0 {
			rh.afterLoadMessages = afterLoadMessages
		}
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
