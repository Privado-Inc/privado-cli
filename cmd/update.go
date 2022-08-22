/**
 * This file is part of Privado OSS.
 *
 * Privado is an open source static code analysis tool to discover data flows in the code.
 * Copyright (C) 2022 Privado, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * For more information, contact support@privado.ai
 *
 */

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/Privado-Inc/privado-cli/pkg/fileutils"
	"github.com/Privado-Inc/privado-cli/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for latest release and update to the latest version Privado CLI",
	Long:  "Check for latest release and update to the latest version Privado CLI",
	Args:  cobra.ExactArgs(0),
	Run:   update,
}

func checkForUpdate() (hasUpdate bool, updateMessage string, err error) {
	if Version == "dev" {
		return false, "", nil
	}

	// get release info (nil when not available)
	releaseInfo, err := utils.GetLatestReleaseFromGitHub(config.AppConfig.PrivadoRepositoryName)

	if err != nil || releaseInfo == nil || releaseInfo.TagName == "" || releaseInfo.PublishedAt == "" {
		return false, "", err
	}

	// compare release, -1, 0, 1
	if semver.Compare(releaseInfo.TagName, Version) > 0 {
		hasUpdate = true
		// Get new release information (with time elapsed if possible)
		daysSinceRelease, err := utils.GetDaysSinceRFC3339String(releaseInfo.PublishedAt)
		if err != nil {
			updateMessage = fmt.Sprintf("New release found: %s\n", releaseInfo.TagName)
		} else {
			daySinceString := ""
			switch {
			case (daysSinceRelease < 1):
				daySinceString = "Released today"
			case daysSinceRelease == 1:
				daySinceString = "Released yesterday"
			default:
				daySinceString = fmt.Sprintf("Released %d days ago", daysSinceRelease)
			}
			updateMessage = fmt.Sprintf("New release found: %s (%s)", releaseInfo.TagName, daySinceString)
		}
	}

	return hasUpdate, updateMessage, nil
}

func update(cmd *cobra.Command, args []string) {
	version(cmd, args)
	fmt.Println()
	time.Sleep(config.AppConfig.SlowdownTime)
	if Version == "dev" {
		exit(
			fmt.Sprint("Cannot perform an update on the dev build. Kindly use a release build or update manually\nFor more information, visit ", config.AppConfig.PrivadoRepository),
			false,
		)
	}

	// get path to current executable
	currentExecPath, err := fileutils.GetPathToCurrentBinary()
	if err != nil {
		exitUpdate(fmt.Sprint("Could not evaluate path to current binary. Auto update fail\nFor more information, visit", config.AppConfig.PrivadoRepository), true)
	}

	// check for appropriate permissions
	hasPerm, err := fileutils.HasWritePermissionToFileNew(currentExecPath)
	if err != nil {
		exitUpdate(fmt.Sprintf("Could not open executable for write: %s", err), true)
	}
	if !hasPerm {
		fmt.Println("> Error: Permission denied")
		fmt.Printf("> The identified installation (%s) requires privileged permissions\n", currentExecPath)
		fmt.Println()
		exit("Try again with a privileged user (sudo)?", true)
	}

	// check for release info
	fmt.Println("Fetching latest release..")
	hasUpdate, updateMessage, err := checkForUpdate()
	if err != nil {
		exitUpdate("Could not fetch latest release. Some error occurred", true)
	}
	if !hasUpdate {
		exit(fmt.Sprint("You are already using the latest version of Privado CLI: ", Version), false)
	}
	fmt.Println(updateMessage)
	time.Sleep(config.AppConfig.SlowdownTime)

	// get download url
	replacer := strings.NewReplacer(
		"${REPO_NAME}", config.AppConfig.PrivadoRepositoryName,
		"${REPO_TAG}", "latest",
		"${REPO_RELEASE_FILE}", config.AppConfig.PrivadoRepositoryReleaseFilename,
	)
	githubReleaseDownloadURL := replacer.Replace(config.ExtConfig.GitHubReleaseDownloadURL)

	// create temporary directory for update assets
	// another approach is to use installation dir to create temp dir instead of systemm default
	// temporaryDirectory, err := ioutil.TempDir(filepath.Dir(currentExecPath), "privado-update-")
	// that is good in general due to no unexpected permission or partition issues with os.Rename
	// but can cause bad side-effects in case of errors
	temporaryDirectory, err := ioutil.TempDir("", "privado-update-")
	if err != nil {
		exitUpdate("Could not create temporary download file. Terminating..", true)
	}
	defer func() {
		// ignore errors (tmp dir)
		os.RemoveAll(temporaryDirectory)
	}()

	downloadedFilePath := filepath.Join(temporaryDirectory, config.AppConfig.PrivadoRepositoryReleaseFilename)

	// download to file in temp directory
	err = utils.DownloadToFile(githubReleaseDownloadURL, downloadedFilePath)
	if err != nil {
		exitUpdate(fmt.Sprint("Could not download release asset: ", githubReleaseDownloadURL), true)
	}
	fmt.Println()
	fmt.Println("Downloaded release asset:", githubReleaseDownloadURL)
	time.Sleep(config.AppConfig.SlowdownTime)

	// extract .tar.gz
	fmt.Println()
	fmt.Println("Extracting release asset..")
	err = fileutils.ExtractTarGzFile(downloadedFilePath, temporaryDirectory)
	if err != nil {
		exitUpdate(fmt.Sprintf("Could not extract release asset: %s: %v", downloadedFilePath, err), true)
	}

	fmt.Println("Extracted release asset:", temporaryDirectory)
	time.Sleep(config.AppConfig.SlowdownTime)
	fmt.Println()

	// Replace existing binary (in current execution) by the updated binary
	fmt.Println("Installing latest release..")
	time.Sleep(config.AppConfig.SlowdownTime)
	err = fileutils.SafeMoveFile(filepath.Join(temporaryDirectory, "privado"), currentExecPath, true)
	if err != nil {
		exitUpdate(fmt.Sprint("Could not update existing installation: ", err), true)
	}

	// woof! all done.
	time.Sleep(config.AppConfig.SlowdownTime)
	fmt.Println()
	fmt.Println("Installed latest release!")
	fmt.Println("To validate installation, run `privado version`")
}

func exitUpdate(msg string, isError bool) {
	fmt.Println(msg)
	fmt.Println()
	exit(fmt.Sprint("> Auto-update failed. Kindly try again or reinstall to update: ", config.AppConfig.PrivadoRepository), isError)
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
