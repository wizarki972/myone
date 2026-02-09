package fldir

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// PATH

func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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

func WriteFile(content, path string) {
	CreateDirectory(filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), os.ModePerm); err != nil {
		panic(err)
	}
}

func CopyFile(source_path, destination_path string) {
	info, err := os.Stat(source_path)
	if err != nil {
		panic(err)
	}

	source, err := os.Open(source_path)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	CreateDirectory(filepath.Dir(destination_path))
	destination, err := os.OpenFile(destination_path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	if _, err = io.Copy(destination, source); err != nil {
		panic(err)
	}

	if err = os.Chmod(destination_path, info.Mode()); err != nil {
		panic(err)
	}
}

// DIRS

func CreateDirectory(path string) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return
		}

		slog.Error("A file already exists in the path : " + path)
		os.Exit(1)
	}

	if !os.IsNotExist(err) {
		panic(err)
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
}
