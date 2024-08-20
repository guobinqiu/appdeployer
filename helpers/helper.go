package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
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
		dir, _ := os.UserHomeDir()
		return filepath.Join(dir, path[1:])
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

func WriteFile(path string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func AppendFile(path string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, perm)
	if err != nil {
		return fmt.Errorf("failed to open file for appending: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func SetDefault(f interface{}, v interface{}) {
	field := reflect.ValueOf(f).Elem()
	defaultValue := reflect.ValueOf(v)
	if field.IsZero() {
		field.Set(defaultValue)
	}
}

func FindRootDir(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}
	return ""
}
