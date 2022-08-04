package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
)

var AppConfig *Configuration

type Configuration struct {
	HomeDirectory                    string
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
	PrivadoTelemetryEndpoint         string
	SlowdownTime                     time.Duration
	Container                        *ContainerConfiguration
}

type ContainerConfiguration struct {
	ImageURL                string
	DockerAccessKeyEnv      string
	UserKeyVolumeDir        string
	DockerKeyVolumeDir      string
	UserConfigVolumeDir     string
	LogConfigVolumeDir      string
	SourceCodeVolumeDir     string
	InternalRulesVolumeDir  string
	ExternalRulesVolumeDir  string
	M2PackageCacheVolumeDir string
}

// init function for AppConfig
func init() {
	home, _ := homedir.Dir()

	imageTag := "niagara-dev"
	telemetryHost := "cli.privado.ai"

	// if PRIVADO_DEV is set, ise developer env settings
	isDev, _ := strconv.ParseBool(os.Getenv("PRIVADO_DEV"))
	// if the running executable is running from the temp dir
	// consider this to be run using "go run main.go",
	// and use developer env settings
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		isDev = true
	}

	if isDev {
		telemetryHost = "t.cli.privado.ai"
		// if PRIVADO_TAG is set, use the specified cli image tag
		imageTag = os.Getenv("PRIVADO_TAG")
		if imageTag == "" {
			imageTag = "niagara-dev"
		}
	}

	AppConfig = &Configuration{
		HomeDirectory:                    home,
		ConfigurationDirectory:           filepath.Join(home, ".privado"),
		UserConfigurationFilePath:        filepath.Join(home, ".privado", "config.json"),
		UserKeyDirectory:                 filepath.Join(home, ".privado", "keys"),
		UserKeyPath:                      filepath.Join(home, ".privado", "keys", "user.key"),
		M2PackageCacheDirectory:          filepath.Join(home, ".m2"),
		PrivacyResultsPathSuffix:         filepath.Join(".privado", "privado.json"),
		PrivadoRepository:                "https://github.com/Privado-Inc/privado-cli",
		PrivadoRepositoryName:            "Privado-Inc/privado-cli",
		PrivadoRepositoryReleaseFilename: fmt.Sprintf("privado-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH),
		PrivadoTelemetryEndpoint:         fmt.Sprintf("https://%s/api/event?version=2", telemetryHost),
		SlowdownTime:                     600 * time.Millisecond,
		Container: &ContainerConfiguration{
			ImageURL:                fmt.Sprintf("public.ecr.aws/privado/cli:%s", imageTag),
			DockerAccessKeyEnv:      "PRIVADO_DOCKER_ACCESS_KEY",
			UserKeyVolumeDir:        "/app/keys/user.key",
			DockerKeyVolumeDir:      "/app/keys/docker.key",
			UserConfigVolumeDir:     "/app/config/config.json",
			LogConfigVolumeDir:      "/app/config/log4j2.xml",
			SourceCodeVolumeDir:     "/app/code",
			InternalRulesVolumeDir:  "/app/rules",
			ExternalRulesVolumeDir:  "/app/external-rules",
			M2PackageCacheVolumeDir: "/root/.m2",
		},
	}
}
