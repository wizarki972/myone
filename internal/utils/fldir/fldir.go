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

	"github.com/wizarki972/myone/internal/utils/cmds"
)

// PATHS
// if the path accessible then returns true, else it returns false
func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FILES
// CreateFile creates a file if not exists.
// If already exists, then overwrites(empties) it and return *os.File.
// If a directory is present there then returns PathError.
func CreateFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Reads the file data as string
func ReadFileAsString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Reads the file data as string, if an error occurs then it returns empty string
// remove this after rewriting battery service.
func ReadFileAsStringNoError(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Writes string as a file in the given path
func WriteStringToFile(content, path string) error {
	if err := CreateDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	// 0664 - 6 (r+w), 4(r)
	if err := os.WriteFile(path, []byte(content), 0664); err != nil {
		return err
	}
	return nil
}

// copy file(s) from source to destination
func CopyFile(source_path, destination_path string) error {
	info, err := os.Stat(source_path)
	if err != nil {
		return err
	}

	source, err := os.Open(source_path)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := CreateDirectory(filepath.Dir(destination_path)); err != nil {
		return err
	}
	destination, err := os.OpenFile(destination_path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err = io.Copy(destination, source); err != nil {
		return err
	}

	if err = os.Chmod(destination_path, info.Mode()); err != nil {
		return err
	}
	return nil
}

// DIRS
// Creates directory
func CreateDirectory(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return nil
		}

		return errors.New("a file already exists in the path : " + path)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

// ARCHIVES
// Extracts .zip files to destination
func Unzip(source, destination string) error {
	zip_file, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer zip_file.Close()

	for _, file := range zip_file.File {
		target := filepath.Join(destination, file.Name)

		// Zip slip protection
		if !strings.HasPrefix(target, filepath.Clean(destination)+string(os.PathSeparator)) {
			return errors.New("illegal path: " + target)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(out, reader)
		reader.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// URLs
// Reads text from URL
func ReadTextFileFromURL(URL string, save bool, save_path string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("http error: " + resp.Status)
	}

	if save {
		out, err := os.OpenFile(save_path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return "", err
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err != nil {
			return "", err
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
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

// Downloads files from the given URL
func DownloadURL(URL, destination string, want_progress bool) error {
	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(errors.New("http error: " + resp.Status))
	}

	if err := CreateDirectory(filepath.Dir(destination)); err != nil {
		return err
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	if want_progress {
		pw := &progress_writter{Total: resp.ContentLength}
		if _, err = io.Copy(out, io.TeeReader(resp.Body, pw)); err != nil {
			fmt.Println()
			panic(err)
		}
		fmt.Println("\r-> Finished Downloading")
	} else if _, err := io.Copy(out, resp.Body); err != nil {
		panic(err)
	}
	return nil
}

// USER
// Only for these two funcs error is not returned, since it is not a expected behaviour.
// I AM JUST LAZY TO RETURN ERROR FRO THESE...

// Returns XDG directory of the input...
func GetXDGDir(name string) string {
	output, err := cmds.ExecCommand("xdg-user-dir "+name, false, true)
	if err != nil {
		panic(err)
	}
	return string(output)
}

// Returns user's home directroy path
func GetHomeDir() string {
	out, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return out
}

// MOVE

// Move moves a file or directory from src to dst.
// Works across filesystems and preserves permissions.
func Move(src, dst string) error {
	// Try fast path (rename)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	slog.Warn("Fast renaming did not work, falling back to moving file by file.")
	if info.IsDir() {
		if err := moveDir(src, dst); err != nil {
			return err
		}
	}

	if err := moveFile(src, dst, info); err != nil {
		return err
	}
	return nil
}

// Moves a file, but not a directory
func moveFile(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	if err = out.Sync(); err != nil {
		return err
	}

	return os.Remove(src)
}

// Moves the directory
func moveDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		if err := moveFile(path, target, info); err != nil {
			return err
		}
		return nil
	})
}
