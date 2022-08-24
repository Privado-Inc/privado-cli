/**
 * This file is part of Privado OSS.
 *
 * Privado is an open source static code analysis tool to discover data flows in the code.
 * Copyright (C) 2022 Privado, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * For more information, contact support@privado.ai
 */

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

	"github.com/Privado-Inc/privado-cli/pkg/config"
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
