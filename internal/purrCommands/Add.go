package purrCommands

import (
	"Persephone/internal/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func AddPurrFiles(arg ...string) error {
	// Case: if `purr init` was not done before, gives an error
	targetDir := filepath.Join(".", ".purr")
	ok, err := utils.ExistsAndIsDirectory(targetDir)
	if err != nil || !ok {
		fmt.Printf("Error: .purr directory not initialized — %v\n", err)
		os.Exit(1)
	}
	// Get Current Working Directory
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Unable to get Working Directory %s \n", err)
		return nil
	}

	// Case: only `purr add` was written
	if len(arg) == 0 {
		fmt.Println("No Files added")
		return nil
	}

	//Detect if the user passed . (all files) or specific files.
	if len(arg) == 1 && arg[0] == "." {
		addAllPurrFiles(dirPath)
	} else {
		addSpecificPurrFiles(dirPath)
	}
	return nil
}

// Called by func AddPurrFiles() when User passed `purr add .` (all files)
func addAllPurrFiles(path string) error {
	// Load all index entries from .purr/index file to IndexEntries
	IndexEntries, _ := utils.ReadIndex(filepath.Join(path, ".purr", "index"))

	// Create a map for faster lookups (path -> entry)
	indexMap := make(map[string]*utils.IndexEntry)
	for i := range IndexEntries {
		indexMap[IndexEntries[i].Path] = &IndexEntries[i]
	}

	// Use up to 5× CPU cores as worker limit
	numWorkers := runtime.NumCPU() * 5
	semaphore := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	utils.WalkAndAddFiles(path, func(filePath string) error {
		wg.Add(1)
		go func(tempPath string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire slot (blocks if at limit)
			defer func() { <-semaphore }() // Release slot when done

			// Getting file Info
			fileInfo, err := os.Stat(tempPath)
			if err != nil {
				log.Printf("failed to stat %s: %v", tempPath, err)
				return
			}

			// Get relative path from repo root
			relPath, err := filepath.Rel(path, tempPath)
			if err != nil {
				log.Printf("failed to get relative path for %s: %v", tempPath, err)
				return
			}

			// Check if file exists in index
			existingEntry, exists := indexMap[relPath]
			if exists {
				if fileInfo.ModTime().Equal(existingEntry.Mtime) {
					return
				}
			}

			// File is new or modified - write blob
			hash, err := utils.WriteBlobWithSHA(tempPath)
			if err != nil {
				log.Printf("failed to write blob with SHA for %s: %v", tempPath, err)
				return
			}

			// Create new entry with all fields populated
			newEntry := utils.PopulateAllIndexField(fileInfo, relPath)
			newEntry.Sha1 = hash

			// Update map with lock
			mu.Lock()
			indexMap[relPath] = &newEntry
			mu.Unlock()

		}(filePath)
		return nil
	})

	// Wait for all goroutines to finish
	wg.Wait()

	// Convert map to slice after all updates are complete
	var updatedEntries []utils.IndexEntry
	for _, entry := range indexMap {
		updatedEntries = append(updatedEntries, *entry)
	}

	// Write updated index to disk
	indexPath := filepath.Join(path, ".purr", "index")
	if err := utils.WriteIndex(indexPath, updatedEntries); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// Called by func AddPurrFiles() when User passed `purr add file1 ...` (specific files)
func addSpecificPurrFiles(path string) {

}
