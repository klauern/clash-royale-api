package datapath

import (
	"fmt"
	"os"
	"path/filepath"
)

const AppDirName = ".cr-api"

func AppDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, AppDirName), nil
}

func AppDirOrFallback() string {
	appDir, err := AppDir()
	if err != nil {
		return filepath.Join(".", AppDirName)
	}
	return appDir
}

func AppPath(parts ...string) (string, error) {
	appDir, err := AppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{appDir}, parts...)...), nil
}

func AppPathOrFallback(parts ...string) string {
	return filepath.Join(append([]string{AppDirOrFallback()}, parts...)...)
}

func DiscoveryDir() (string, error) {
	return AppPath("discover")
}

func DiscoveryDirOrFallback() string {
	return AppPathOrFallback("discover")
}

func LeaderboardsDir() (string, error) {
	return AppPath("leaderboards")
}

func LeaderboardDBPath(sanitizedTag string) (string, error) {
	leaderboardsDir, err := LeaderboardsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(leaderboardsDir, fmt.Sprintf("%s.db", sanitizedTag)), nil
}

func FuzzStorageDBPath(defaultDBName string) (string, error) {
	return AppPath(defaultDBName)
}
