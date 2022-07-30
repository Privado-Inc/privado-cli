package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/mitchellh/go-homedir"
)

var AppConfig *Configuration

type Configuration struct {
	HomeDirectory                    string
	DefaultWebPort                   int
	ConfigurationDirectory           string
	UserConfigurationFilePath        string
	UserKeyDirectory                 string
	UserKeyPath                      string
	M2PackageCacheDirectory          string
	PrivacyResultsPathSuffix         string
	PrivacyReportsDirectorySuffix    string
	PrivadoRepository                string
	PrivadoRepositoryName            string
	PrivadoRepositoryReleaseFilename string
	SlowdownTime                     time.Duration
	Container                        *ContainerConfiguration
}

type ContainerConfiguration struct {
	ImageURL                string
	UserKeyVolumeDir        string
	DockerKeyVolumeDir      string
	UserConfigVolumeDir     string
	SourceCodeVolumeDir     string
	InternalRulesVolumeDir  string
	ExternalRulesVolumeDir  string
	M2PackageCacheVolumeDir string
}

// init function for AppConfig
func init() {
	home, _ := homedir.Dir()

	imageTag := "niagara-dev"

	if isDev, err := strconv.ParseBool(os.Getenv("PRIVADO_DEV")); err == nil && isDev {
		imageTag = os.Getenv("PRIVADO_TAG")
		if imageTag == "" {
			imageTag = "niagara-dev"
		}
	}

	AppConfig = &Configuration{
		HomeDirectory:                    home,
		DefaultWebPort:                   3000,
		ConfigurationDirectory:           filepath.Join(home, ".privado"),
		UserConfigurationFilePath:        filepath.Join(home, ".privado", "config.json"),
		UserKeyDirectory:                 filepath.Join(home, ".privado", "keys"),
		UserKeyPath:                      filepath.Join(home, ".privado", "keys", "user.key"),
		M2PackageCacheDirectory:          filepath.Join(home, ".m2"),
		PrivacyResultsPathSuffix:         filepath.Join(".privado", "privacy.json"),
		PrivadoRepository:                "https://github.com/Privado-Inc/privado-cli",
		PrivadoRepositoryName:            "Privado-Inc/privado-cli",
		PrivadoRepositoryReleaseFilename: fmt.Sprintf("privado-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH),
		SlowdownTime:                     600 * time.Millisecond,
		Container: &ContainerConfiguration{
			ImageURL:                fmt.Sprintf("public.ecr.aws/privado/cli:%s", imageTag),
			UserKeyVolumeDir:        "/app/keys/user.key",
			DockerKeyVolumeDir:      "/app/keys/docker.key",
			UserConfigVolumeDir:     "/app/config/config.json",
			SourceCodeVolumeDir:     "/app/code",
			InternalRulesVolumeDir:  "/app/rules",
			ExternalRulesVolumeDir:  "/app/external-rules",
			M2PackageCacheVolumeDir: "/root/.m2",
		},
	}
}
