package fldir

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// rename this func to WriteStringtTFile
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

// ARCHIVES

func Unzip(source, destination string) {
	zip_file, err := zip.OpenReader(source)
	if err != nil {
		panic(err)
	}
	defer zip_file.Close()

	for _, file := range zip_file.File {
		target := filepath.Join(destination, file.Name)

		// Zip slip protection
		if !strings.HasPrefix(target, filepath.Clean(destination)+string(os.PathSeparator)) {
			panic(errors.New("illegal path: " + target))
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				panic(err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			panic(err)
		}

		reader, err := file.Open()
		if err != nil {
			panic(err)
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(out, reader)
		reader.Close()
		out.Close()
		if err != nil {
			panic(err)
		}
	}
}

// URLs

func ReadTextFileFromURL(URL string, save bool, save_path string) string {
	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(errors.New("http error: " + resp.Status))
	}

	if save {
		out, err := os.OpenFile(save_path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err != nil {
			panic(err)
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(data)
}

type progress_writter struct {
	Total   int64
	Written int64
}

func (pw *progress_writter) Write(b []byte) (int, error) {
	n := len(b)
	pw.Written += int64(n)
	percent := float64(pw.Written) / float64(pw.Total) * 100
	fmt.Printf("\rDownloading... %.3f%%", percent)
	return n, nil
}

func DownloadURL(URL, destination string) {
	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(errors.New("http error: " + resp.Status))
	}

	CreateDirectory(filepath.Dir(destination))
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	pw := &progress_writter{Total: resp.ContentLength}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, pw)); err != nil {
		fmt.Println()
		panic(err)
	}
	fmt.Println("\nDone")
}
