package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Privado-Inc/privado/pkg/config"
	"github.com/schollz/progressbar/v3"
)

type gitHubReleaseType struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
}

func GetLatestReleaseFromGitHub(repoName string) (*gitHubReleaseType, error) {
	endpoint := strings.Replace(config.ExtConfig.GitHubReleasesEndpoint, "${REPO_NAME}", config.AppConfig.PrivadoRepositoryName, 1)
	url := fmt.Sprintf("%s%s", config.ExtConfig.GitHubAPIHost, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	defaultClient := &http.Client{}
	response, err := defaultClient.Do(req)

	if err != nil || response.StatusCode != 200 {
		return nil, err
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	releaseResponse := gitHubReleaseType{}
	err = json.Unmarshal(responseData, &releaseResponse)
	if err != nil {
		return nil, err
	}

	return &releaseResponse, nil
}

func DownloadToFile(downloadURL, filePath string) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading..",
	)

	_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetDaysSinceRFC3339String(date string) (int, error) {
	parsedDate, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return 0, err
	}
	elapsedTime := time.Since(parsedDate)
	elapsedDays := int(math.Round(elapsedTime.Hours() / 24))
	return elapsedDays, nil
}
