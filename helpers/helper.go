package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func GetDefaultKubeConfig() string {
	if !IsBlank(os.Getenv("KUBECONFIG")) {
		return os.Getenv("KUBECONFIG")
	}
	return filepath.Join(os.Getenv("HOME"), ".kube", "config")
}

func GetDefaultDockerConfig() string {
	return filepath.Join(os.Getenv("HOME"), ".docker", "config.json")
}

func IsFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func IsDirExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func ExpandUser(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[1:])
	}
	return path
}

func ListSubDirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var subDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			subDirs = append(subDirs, entry.Name())
		}
	}
	return subDirs, nil
}

func Contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}
