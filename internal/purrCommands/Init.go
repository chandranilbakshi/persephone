package purrCommands

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

func InitPurrDirectories(basePath string) error {
	purrDir := filepath.Join(basePath, ".purr")

	dirs := []string{
		filepath.Join(purrDir, "objects"),
		filepath.Join(purrDir, "refs", "heads"),
		filepath.Join(purrDir, "logs"),
	}

	// Create all directories
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Set hidden attribute on Windows
	if runtime.GOOS == "windows" {
		purrDirPtr, err := syscall.UTF16PtrFromString(purrDir)
		if err != nil {
			return err
		}

		// Get current attributes
		attrs, err := syscall.GetFileAttributes(purrDirPtr)
		if err != nil {
			return err
		}

		// Add hidden attribute (FILE_ATTRIBUTE_HIDDEN = 0x2)
		err = syscall.SetFileAttributes(purrDirPtr, attrs|syscall.FILE_ATTRIBUTE_HIDDEN)
		if err != nil {
			return err
		}
	}

	return nil
}
