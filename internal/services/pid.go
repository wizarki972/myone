package services

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

func savePID(fileName string, pid int) error {
	path := filepath.Join(common.RUN_DIR, strconv.Itoa(os.Getuid()), fileName)
	return fldir.WriteStringToFile(strconv.Itoa(pid), path)
}

func isOldProcessRunning(fileName string) (bool, int, error) {
	path := filepath.Join(common.RUN_DIR, strconv.Itoa(os.Getuid()), fileName)
	if !fldir.IsPathExist(path) {
		return false, -1, nil
	}

	pidStr, err := fldir.ReadFileAsString(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, -1, nil
		}
		return false, -1, err
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false, -1, err
	}

	process, err := os.FindProcess(pid)
	err = process.Signal(syscall.Signal(0))
	if err == nil || errors.Is(err, os.ErrPermission) {
		return true, pid, nil
	}
	return false, -1, nil
}

func getPID(fileName string) (int, error) {
	path := filepath.Join(common.RUN_DIR, strconv.Itoa(os.Getuid()), fileName)
	pid, err := fldir.ReadFileAsString(path)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(pid)
}

func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err = process.Kill(); err != nil {
		return err
	}
	return nil
}
