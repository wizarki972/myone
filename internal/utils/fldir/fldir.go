package fldir

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// STATIC VALUES
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return home
}

// FILES
func CreateFile(path string) *os.File {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		panic(err)
	}

	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	return file
}

func ReadFileAsString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadFileAsStringNoError(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// DIRS

func GetDirEntry(path, entry_substr string) (os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), entry_substr) {
			return entry, nil
		}
	}

	return nil, errors.New("dir entry not found")
}
