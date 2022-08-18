package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Privado-Inc/privado-cli/pkg/fileutils"
	"github.com/mitchellh/go-homedir"
)

var AppConfig *Configuration

type Configuration struct {
	HomeDirectory                    string
	CacheDirectory                   string
	ConfigurationDirectory           string
	UserConfigurationFilePath        string
	UserKeyDirectory                 string
	UserKeyPath                      string
	M2CacheDirectoryName             string
	GradleCacheDirectoryName         string
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
	ImageURL                    string
	DockerAccessKeyEnv          string
	UserKeyVolumeDir            string
	DockerKeyVolumeDir          string
	UserConfigVolumeDir         string
	LogConfigVolumeDir          string
	SourceCodeVolumeDir         string
	InternalRulesVolumeDir      string
	ExternalRulesVolumeDir      string
	M2PackageCacheVolumeDir     string
	GradlePackageCacheVolumeDir string
}

// init function for AppConfig
func init() {
	home, _ := homedir.Dir()

	imageTag := "niagara-dev"
	telemetryHost := "t.cli.privado.ai"

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
		M2CacheDirectoryName:             ".m2",
		GradleCacheDirectoryName:         ".gradle",
		PrivacyResultsPathSuffix:         filepath.Join(".privado", "privado.json"),
		PrivadoRepository:                "https://github.com/Privado-Inc/privado-cli",
		PrivadoRepositoryName:            "Privado-Inc/privado-cli",
		PrivadoRepositoryReleaseFilename: fmt.Sprintf("privado-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH),
		PrivadoTelemetryEndpoint:         fmt.Sprintf("https://%s/api/event?version=2", telemetryHost),
		SlowdownTime:                     600 * time.Millisecond,
		Container: &ContainerConfiguration{
			ImageURL:                    fmt.Sprintf("public.ecr.aws/privado/cli:%s", imageTag),
			DockerAccessKeyEnv:          "PRIVADO_DOCKER_ACCESS_KEY",
			UserKeyVolumeDir:            "/app/keys/user.key",
			DockerKeyVolumeDir:          "/app/keys/docker.key",
			UserConfigVolumeDir:         "/app/config/config.json",
			LogConfigVolumeDir:          "/app/config/log4j2.xml",
			SourceCodeVolumeDir:         "/app/code",
			InternalRulesVolumeDir:      "/app/rules",
			ExternalRulesVolumeDir:      "/app/external-rules",
			M2PackageCacheVolumeDir:     "/root/.m2",
			GradlePackageCacheVolumeDir: "/root/.gradle",
		},
	}

	privadoCacheDir, _ := initPrivadoCacheDirectory()
	AppConfig.CacheDirectory = privadoCacheDir
}

// returns existing privado cache directory
// if not available - creates one and returns
func initPrivadoCacheDirectory() (string, error) {
	cacheDir := getPrivadoCacheDirectory()
	if cacheDir != "" {
		return cacheDir, nil
	}
	return createPrivadoCacheDirectory()
}

func createPrivadoCacheDirectory() (string, error) {
	if systemDefinedCacheDir, err := os.UserCacheDir(); err != nil {
		location := filepath.Join(AppConfig.ConfigurationDirectory, ".cache")
		if err := os.MkdirAll(location, os.ModePerm); err != nil {
			return "", err
		}
		return location, nil
	} else {
		location := filepath.Join(systemDefinedCacheDir, "privado")
		if err := os.MkdirAll(location, os.ModePerm); err != nil {
			return "", err
		}
		return location, nil
	}
}

// Opposite direction from create - check if fallbacks are created first
// then going forward, continue to use them instead of creating other dir
func getPrivadoCacheDirectory() string {
	location := filepath.Join(AppConfig.ConfigurationDirectory, ".cache")
	if exists, _ := fileutils.DoesFileExists(location); exists {
		return location
	}

	if systemDefinedCacheDir, err := os.UserCacheDir(); err == nil {
		location := filepath.Join(systemDefinedCacheDir, "privado")
		if exists, _ := fileutils.DoesFileExists(location); exists {
			return location
		}
	}

	return ""
}

func GetPackageCacheDirectory(packageManager string) (string, error) {
	var packageCacheDir string
	switch packageManager {
	case "m2":
		packageCacheDir = AppConfig.M2CacheDirectoryName
	case "gradle":
		packageCacheDir = AppConfig.GradleCacheDirectoryName
	default:
		packageCacheDir = AppConfig.GradleCacheDirectoryName
	}

	cacheDir := AppConfig.CacheDirectory
	if cacheDir != "" {
		if exists, err := fileutils.DoesFileExists(filepath.Join(cacheDir, packageCacheDir)); err != nil {
			return "", err
		} else if exists {
			return filepath.Join(cacheDir, packageCacheDir), nil
		}
	}

	home, _ := homedir.Dir()
	defaultPackageCacheLocation := filepath.Join(home, packageCacheDir)
	if exists, err := fileutils.DoesFileExists(defaultPackageCacheLocation); err != nil {
		return "", err
	} else if exists {
		// if default package location exists, use that (~/.m2, ~/.gradle)
		return defaultPackageCacheLocation, nil
	} else {
		// if default location does not exist, create dir in PrivadoCache and use that one
		// if cacheDir is empty, try creating again
		if cacheDir == "" {
			cacheDir, err = createPrivadoCacheDirectory()
			if err != nil {
				return "", err
			}
		}

		location := filepath.Join(cacheDir, packageCacheDir)
		if err := os.MkdirAll(location, os.ModePerm); err != nil {
			return "", err
		}

		return location, nil
	}
}
