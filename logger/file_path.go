package logger

import (
	"os"
	"path/filepath"
)

func CheckPathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func CreatePath(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func CheckAndCreate(path string) error {
	if !CheckPathExists(path) {
		return CreatePath(path)
	}
	return nil
}

func GetDirFromPath(path string) string {
	return filepath.Dir(path)
}

func GetFileNameFromPath(path string) string {
	_, f := filepath.Split(path)
	return f
}
