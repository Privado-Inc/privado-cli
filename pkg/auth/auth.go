package auth

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"

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

// Returns a string UUID
func GenerateUserKey() string {
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
	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// parse uuid
	if _, err = uuid.ParseBytes(dataBytes); err != nil {
		return err
	}

	return nil
}
