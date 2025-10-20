package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
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

/*
WriteBlobWithSHA creates a Git-style blob object from a file and stores it in .purr/objects.
It reads the file, prepends a "blob <size>\x00" header, computes the SHA-1 hash of the
combined data, and writes the zlib-compressed blob to .purr/objects/xx/yyyy... where xx
is the first 2 characters of the hash and yyyy... is the remaining hash. This matches
Git's object storage format, allowing content-addressable storage where the hash serves
as the unique identifier for the file's contents.
*/
func WriteBlobWithSHA(filePath string) ([20]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return [20]byte{}, err
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

	// Call helper to store object
	err = storeObject(hashStr, compressed.Bytes())
	if err != nil {
		return [20]byte{}, err
	}

	return hash, nil
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

/*
ReadIndex reads and deserializes the .purr/index file.
Git index format: 12-byte header + repeated entries.
Each entry: fixed metadata (62 bytes) + variable-length path + padding.
*/
func ReadIndex(indexPath string) ([]IndexEntry, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	buf := bytes.NewReader(data) // creates a readable buffer from a byte slice
	var entries []IndexEntry

	// Skip 12-byte header
	if _, err := buf.Seek(12, io.SeekStart); err != nil {
		return nil, fmt.Errorf("invalid index header: %w", err)
	}

	for buf.Len() > 0 {
		var entry IndexEntry

		// Reading fixed 62-byte metadata block from the memory buffer (buf) and storing its value into different variables
		// Read IndexEntry struct to better understand it
		// binary.BigEndian defines byte order: most significant byte comes first

		var ctime, mtime int64
		if err := binary.Read(buf, binary.BigEndian, &ctime); err != nil {
			return nil, fmt.Errorf("failed to read ctime: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &mtime); err != nil {
			return nil, fmt.Errorf("failed to read mtime: %w", err)
		}
		entry.Ctime = time.Unix(ctime, 0)
		entry.Mtime = time.Unix(mtime, 0)

		if err := binary.Read(buf, binary.BigEndian, &entry.Dev); err != nil {
			return nil, fmt.Errorf("failed to read dev: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Ino); err != nil {
			return nil, fmt.Errorf("failed to read ino: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Mode); err != nil {
			return nil, fmt.Errorf("failed to read mode: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Uid); err != nil {
			return nil, fmt.Errorf("failed to read uid: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Gid); err != nil {
			return nil, fmt.Errorf("failed to read gid: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Size); err != nil {
			return nil, fmt.Errorf("failed to read size: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Sha1); err != nil {
			return nil, fmt.Errorf("failed to read sha1: %w", err)
		}
		if err := binary.Read(buf, binary.BigEndian, &entry.Stage); err != nil {
			return nil, fmt.Errorf("failed to read stage: %w", err)
		}

		/*
			The padding calculation and subsequent `buf.Seek` ensures that each IndexEntry starts at an 8-byte aligned offset in the index file, as required by the Purr index format.
		*/
		var pathLen uint16
		if err := binary.Read(buf, binary.BigEndian, &pathLen); err != nil {
			return nil, fmt.Errorf("failed to read path length: %w", err)
		}

		pathBytes := make([]byte, pathLen)
		if n, err := buf.Read(pathBytes); err != nil || n != int(pathLen) {
			return nil, fmt.Errorf("failed to read path (expected %d bytes): %w", pathLen, err)
		}
		entry.Path = string(pathBytes)

		// Skip padding to align next entry to 8 bytes
		// Total entry size = 62 (metadata) + 2 (path length) + pathLen (path data)
		entrySize := 62 + 2 + pathLen
		paddingLen := (8 - (entrySize % 8)) % 8
		if _, err := buf.Seek(int64(paddingLen), io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip padding: %w", err)
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

// WriteIndex serializes index entries and writes them to disk in Git index format.
// Format: 12-byte header (DIRC + version 2 + entry count) + entries with 8-byte padding.
func WriteIndex(indexPath string, entries []IndexEntry) error {
	var buf bytes.Buffer

	// Write 12-byte header
	buf.WriteString("DIRC")                                    // Magic signature (4 bytes)
	binary.Write(&buf, binary.BigEndian, uint32(2))            // Version 2 (4 bytes)
	binary.Write(&buf, binary.BigEndian, uint32(len(entries))) // Entry count (4 bytes)

	// Write each entry
	for _, entry := range entries {
		// Write fixed 62-byte metadata
		binary.Write(&buf, binary.BigEndian, entry.Ctime.Unix()) // 8 bytes
		binary.Write(&buf, binary.BigEndian, entry.Mtime.Unix()) // 8 bytes
		binary.Write(&buf, binary.BigEndian, entry.Dev)          // 4 bytes
		binary.Write(&buf, binary.BigEndian, entry.Ino)          // 4 bytes
		binary.Write(&buf, binary.BigEndian, entry.Mode)         // 4 bytes
		binary.Write(&buf, binary.BigEndian, entry.Uid)          // 4 bytes
		binary.Write(&buf, binary.BigEndian, entry.Gid)          // 4 bytes
		binary.Write(&buf, binary.BigEndian, entry.Size)         // 4 bytes
		binary.Write(&buf, binary.BigEndian, entry.Sha1)         // 20 bytes
		binary.Write(&buf, binary.BigEndian, entry.Stage)        // 2 bytes

		// Write path length and path data
		pathBytes := []byte(entry.Path)
		binary.Write(&buf, binary.BigEndian, uint16(len(pathBytes))) // 2 bytes
		buf.Write(pathBytes)

		// Add padding to align to 8-byte boundary
		entrySize := 62 + 2 + len(pathBytes)
		paddingLen := (8 - (entrySize % 8)) % 8
		buf.Write(make([]byte, paddingLen))
	}

	// Write to disk
	if err := os.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}
