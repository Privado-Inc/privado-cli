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
 */

package fileutils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codeclysm/extract/v3"
)

func CopyFile(src, dst string) error {
	// open src
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// create dst
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// copy content
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// get src permissions
	fileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	// give same permissions
	if err := os.Chmod(dst, fileStat.Mode()); err != nil {
		return err
	}

	return out.Close()
}

func DoesFileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func GetAbsolutePath(relativePath string) string {

	fullPath, err := filepath.Abs(relativePath)
	if err != nil {
		panic(err)
	}
	return fullPath
}

func GetPathToCurrentBinary() (string, error) {
	currentFilePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	// resolve any symlinks
	resolvedFilePath, err := filepath.EvalSymlinks(currentFilePath)
	if err != nil {
		return "", err
	}

	return resolvedFilePath, nil
}

func ExtractTarGzFile(sourceFile, target string) error {
	ctx := context.Background()
	data, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(data)
	err = extract.Gz(ctx, buffer, target, func(s string) string { return s })
	if err != nil {
		return err
	}

	return nil
}

// os.Rename panics with "invalid cross-device link"
// for cases where the source is on a different volume
// or if the source volume is masked like in some
// linux systems and hence we attempt to perform
// copy+delete to replicate a move operation
// with added backup fallbacks
func SafeMoveFile(source, target string, showLogs bool) (err error) {
	// Resolve symlinks and use actual path for critical operation
	if source, err = filepath.EvalSymlinks(source); err != nil {
		return err
	}
	if target, err = filepath.EvalSymlinks(target); err != nil {
		return err
	}

	fileExists, err := DoesFileExists(target)
	if err != nil {
		return err
	}

	// if file exists, make a backup and remove
	// file will always exist in case of updates, check to support generic usage
	if fileExists {
		backupTargetFile := filepath.Join(filepath.Dir(source), fmt.Sprintf("%s-backup", filepath.Base(source)))
		if showLogs {
			fmt.Printf("> Creating backup of existing file (%s)\n", backupTargetFile)
		}
		if err = CopyFile(target, backupTargetFile); err != nil {
			return err
		}

		if err = os.Remove(target); err != nil {
			return err
		}

		defer func() {
			// remove backup: when no panic
			if panicErr := recover(); panicErr != nil {
				err = fmt.Errorf("failed to move file, failed to restore backup: %v", panicErr)
			} else {
				if showLogs {
					fmt.Println("> Removing backup file")
				}
				os.Remove(backupTargetFile)
			}
		}()
	}

	err = CopyFile(source, target)
	if err != nil {
		// if file existed initially, place it back
		if showLogs {
			fmt.Println("> Failed to move file:\n", err)
		}

		if fileExists {
			if showLogs {
				fmt.Println("> Restoring from backup")
			}

			// backupTargetFile := filepath.Dir(source) + filepath.Base(source) + "-backup"
			backupTargetFile := filepath.Join(filepath.Dir(source), fmt.Sprintf("%s-backup", filepath.Base(source)))
			if backupErr := CopyFile(backupTargetFile, target); backupErr != nil {
				fmt.Println()
				fmt.Println("Unable to restore original file:\n", backupErr)

				fmt.Println("-------------------")
				fmt.Println("> Kindly move manually to restore original state:")
				fmt.Printf("mv %s %s\n", backupTargetFile, target)
				fmt.Println("-------------------")

				fmt.Println()
				// panic to skip deletion of
				panic(backupTargetFile)
			}
		}
		return err
	}
	if showLogs {
		fmt.Println("> Move successful")
	}

	return nil
}
