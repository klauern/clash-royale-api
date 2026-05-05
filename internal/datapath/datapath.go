package datapath

import (
	"fmt"
	"os"
	"path/filepath"
)

const appDirName = ".cr-api"

func AppDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, appDirName), nil
}

func LeaderboardsDir() (string, error) {
	appDir, err := AppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDir, "leaderboards"), nil
}

func LeaderboardDBPath(sanitizedTag string) (string, error) {
	leaderboardsDir, err := LeaderboardsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(leaderboardsDir, fmt.Sprintf("%s.db", sanitizedTag)), nil
}

func FuzzStorageDBPath(defaultDBName string) (string, error) {
	appDir, err := AppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appDir, defaultDBName), nil
}
