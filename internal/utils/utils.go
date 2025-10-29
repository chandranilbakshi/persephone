package utils

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ExistsAndIsDirectory checks if the given path exists and is present in the directory.
func ExistsAndIsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)

	// Case 1: Path does not exist.
	if os.IsNotExist(err) {
		return false, err
	}

	// Case 2: Error Handling
	if err != nil {
		return false, fmt.Errorf("stat check failed for %s: %w", path, err)
	}

	// Case 3: Path Exists
	return info.IsDir(), nil
}

/*
WalkDir(root, fn) walks through all files and directories under root.
It calls fn(path, d, err) for each entry, including root itself.
The callback can handle each item or skip directories (e.g. return filepath.SkipDir).
Returns an error if the walk fails.
*/
func WalkAndAddFiles(root string, handleFile func(string) error) error {
	return filepath.WalkDir(root, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %w", entryPath, err)
		}

		// Skip anything starting with `.`
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip remaining directories
		if d.IsDir() {
			return nil
		}

		// Handle file, but continue on error
		if err := handleFile(entryPath); err != nil {
			log.Printf("error handling file %s: %v", entryPath, err)
			return nil
		}
		return nil
	})
}

// storeObject handles creating directories and writing compressed blob
func storeObject(hashStr string, data []byte) error {
	dir := filepath.Join(".purr", "objects", hashStr[:2])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	objectPath := filepath.Join(dir, hashStr[2:])
	return os.WriteFile(objectPath, data, 0644)
}

/*
*
PopulateAllIndexField creates an IndexEntry from the provided os.FileInfo and relative path.
It extracts file metadata, handling Windows-specific fields, and populates all index fields.
*/
func PopulateAllIndexField(fileInfo os.FileInfo, relPath string) IndexEntry {
	stat := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	return IndexEntry{
		Ctime: time.Unix(0, stat.CreationTime.Nanoseconds()),
		Mtime: fileInfo.ModTime(),
		Dev:   0, // Not applicable on Windows
		Ino:   0, // Not applicable on Windows
		Mode:  uint32(fileInfo.Mode()),
		Uid:   0, // Not applicable on Windows
		Gid:   0, // Not applicable on Windows
		Size:  uint32(fileInfo.Size()),
		Stage: 0,
		Path:  relPath,
	}
}
