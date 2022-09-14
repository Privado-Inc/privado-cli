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

package auth

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/Privado-Inc/privado-cli/pkg/ci"
	"github.com/Privado-Inc/privado-cli/pkg/fileutils"
	"github.com/google/uuid"
)

// Ensures a validated UserKey exists
func BootstrapUserKey(userKeyPath, userKeyDirectory string) error {
	if keyExists, _ := fileutils.DoesFileExists(userKeyPath); keyExists {
		// if verification fails, continue to regenerate
		if err := VerifyUserKeyFile(userKeyPath); err == nil {
			return nil
		}
	}

	if err := os.MkdirAll(userKeyDirectory, os.ModePerm); err != nil {
		return err
	}

	userKey := GenerateUserKey()
	if err := os.WriteFile(userKeyPath, []byte(userKey), 0600); err != nil {
		return err
	}

	return nil
}

// the fn hashes the supplied string first and uses first 16 bytes
// of the generated hash to be converted to UUID. As each bit of
// SHA hash is effectively random, we get a unique-uuid for each
// distinct string, even when they have minimal differences
func GenerateUserKeyFromString(msg string) string {
	hash := CalculateSHA256HashInBytes(msg)
	uuid, err := uuid.FromBytes(hash[:16])
	if err != nil {
		panic(fmt.Errorf("cannot generate uuid from string %s: %v", msg, err))
	}

	return uuid.String()
}

// Returns a string UUID
func GenerateUserKey() string {
	if ci.CISessionConfig.IsCI {
		// if UserIdentifier is Null for CI Env
		// either we do not support the provider yet, or we were
		// not able to get appropriate user identifier for the
		// provider, in both cases, use a default user id for
		// CI, so all miscellaneous CI runs can be tracked
		// as part of the default identified CI user
		if ci.CISessionConfig.UserIdentifier == "" {
			fmt.Println("> Unknown CI identifier. Setting default CI user")
			ci.CISessionConfig.UserIdentifier = "PrivadoDefaultCIUserIdentifier"
		}

		fmt.Println("> Identified CI user:", ci.CISessionConfig.UserIdentifier)
		fmt.Println()

		return GenerateUserKeyFromString(ci.CISessionConfig.UserIdentifier)
	}

	return uuid.NewString()
}

// Get the user key
func GetUserKey(userKeyPath string) string {
	fileContent, err := os.ReadFile(userKeyPath)
	if err != nil {
		return ""
	}
	id, err := uuid.ParseBytes(fileContent)
	if err != nil {
		return ""
	}

	return id.String()
}

// Calculate user hash from key
func CalculateSHA256Hash(key string) string {
	if key == "" {
		panic("fatal: the function restricts generating hash for empty key string")
	}
	hashByteArray := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hashByteArray[:])
}

// similar to CalculateSHA256Hash, but does not convert byte array to string
// we noticed unwanted type conversions when, converting back to []byte from string
// created another for simplicity and to avoid unnecessary complications
func CalculateSHA256HashInBytes(key string) [32]byte {
	if key == "" {
		panic("fatal: the function restricts generating hash for empty key string")
	}
	return sha256.Sum256([]byte(key))
}

// Gets the user key and returns the calculated hash for it
func GetUserHash(userKeyPath string) string {
	return CalculateSHA256Hash(GetUserKey(userKeyPath))
}

func VerifyUserKeyFile(pathToFile string) error {
	// open file
	file, err := os.Open(pathToFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// read file
	dataBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// parse uuid
	if _, err = uuid.ParseBytes(dataBytes); err != nil {
		return err
	}

	return nil
}
