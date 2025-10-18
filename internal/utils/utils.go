package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
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
Old Sequencial Search (Kept for Notalgia)
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

/*
WriteBlobWithSHA creates a Git-style blob object from a file and stores it in .purr/objects.
It reads the file, prepends a "blob <size>\x00" header, computes the SHA-1 hash of the
combined data, and writes the zlib-compressed blob to .purr/objects/xx/yyyy... where xx
is the first 2 characters of the hash and yyyy... is the remaining hash. This matches
Git's object storage format, allowing content-addressable storage where the hash serves
as the unique identifier for the file's contents.
*/
func WriteBlobWithSHA(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	header := fmt.Sprintf("blob %d\x00", len(content)) // `\x00` -> null

	blob := append([]byte(header), content...)

	hash := sha1.Sum(blob)
	hashStr := fmt.Sprintf("%x", hash) // hex string

	// Compress with zlib
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(blob)
	w.Close()

	dir := filepath.Join(".purr", "objects", hashStr[:2]) // making directory with first two Char
	os.MkdirAll(dir, 0755)
	objectPath := filepath.Join(dir, hashStr[2:]) // adding rest Sha1 key inside that directory
	os.WriteFile(objectPath, blob, 0644)
}
