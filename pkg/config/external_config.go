package config

var ExtConfig *ExternalConfiguration

type ExternalConfiguration struct {
	GitHubAPIHost            string
	GitHubReleasesEndpoint   string
	GitHubReleaseDownloadURL string
	PrivadoTelemeryEndpoint  string
}

// init function for ExtConfig
func init() {
	ExtConfig = &ExternalConfiguration{
		GitHubAPIHost:            "https://api.github.com",
		GitHubReleasesEndpoint:   "/repos/${REPO_NAME}/releases/latest",
		GitHubReleaseDownloadURL: "https://github.com/${REPO_NAME}/releases/download/${REPO_TAG}/${REPO_RELEASE_FILE}",
	}
}
