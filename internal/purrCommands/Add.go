package purrCommands

import (
	"Persephone/internal/utils"
	"fmt"
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
func addAllPurrFiles(path string) {
	numWorkers := runtime.NumCPU() * 5 // Limit concurrent workers
	semaphore := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup

	utils.WalkAndAddFiles(path, func(filePath string) error {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire slot (blocks if at limit)
			defer func() { <-semaphore }() // Release slot when done

			utils.WriteBlobWithSHA(filePath)
			// Your file processing logic here
		}(filePath)
		return nil
	})
	// Wait for all goroutines to finish
	wg.Wait()
}

// Called by func AddPurrFiles() when User passed `purr add file1 ...` (specific files)
func addSpecificPurrFiles(path string) {

}
