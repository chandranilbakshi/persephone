package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

/*
BuildTreeObject constructs a Git tree object from a slice of TreeEntries.
It sorts the entries (directories before files, lexicographically), validates each entry,
and encodes them in the Git tree object format:
  - Each entry is "{mode} {name}\0{20-byte SHA-1}"

Returns the raw bytes of the tree object, or an error if validation fails.
*/
func BuildTreeObject(entries []*TreeEntries) ([]byte, error) {
	// Check for empty entries list
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries to create tree for")
	}

	// Sort entries (directories as "name/", files as "name")
	sort.Slice(entries, func(i, j int) bool {
		nameI := entries[i].Name
		nameJ := entries[j].Name
		if entries[i].IsTree {
			nameI += "/"
		}
		if entries[j].IsTree {
			nameJ += "/"
		}
		return nameI < nameJ
	})

	// Build tree content
	var treeContent []byte
	for _, entry := range entries {
		// Validation: mode and name must not be empty
		if entry.Mode == "" || entry.Name == "" {
			return nil, fmt.Errorf("invalid entry: mode and name required (got mode='%s', name='%s')", entry.Mode, entry.Name)
		}
		// Validation: mode must be a valid Git mode
		if entry.Mode != "100644" && entry.Mode != "100755" && entry.Mode != "040000" {
			return nil, fmt.Errorf("invalid mode for entry %s: %s", entry.Name, entry.Mode)
		}
		// {mode} {name}\0
		line := fmt.Sprintf("%s %s\x00", entry.Mode, entry.Name)
		treeContent = append(treeContent, []byte(line)...)
		// 20 raw bytes of SHA-1 (decode hex)
		shaBytes, err := hex.DecodeString(entry.Sha1Hex)
		if err != nil || len(shaBytes) != 20 {
			return nil, fmt.Errorf("invalid SHA-1 for entry %s", entry.Name)
		}
		treeContent = append(treeContent, shaBytes...)
	}

	// Create tree object
	header := fmt.Sprintf("tree %d\x00", len(treeContent))
	treeObj := append([]byte(header), treeContent...)

	return treeObj, nil
}

func BuildCommitObject(treeSHA1, parentSHA1, authorName, authorEmail, message string) ([]byte, error) {
	// Validate inputs
	if treeSHA1 == "" {
		return nil, fmt.Errorf("tree SHA-1 is required")
	}
	if len(treeSHA1) != 40 {
		return nil, fmt.Errorf("invalid tree SHA-1 length: got %d, want 40", len(treeSHA1))
	}
	if authorName == "" || authorEmail == "" {
		return nil, fmt.Errorf("author name and email are required")
	}
	if message == "" {
		return nil, fmt.Errorf("commit message is required")
	}

	// Build commit content
	var commitContent bytes.Buffer

	// Tree line
	commitContent.WriteString(fmt.Sprintf("tree %s\n", treeSHA1))

	// Parent line (optional for first commit)
	if parentSHA1 != "" {
		if len(parentSHA1) != 40 {
			return nil, fmt.Errorf("invalid parent SHA-1 length: got %d, want 40", len(parentSHA1))
		}
		commitContent.WriteString(fmt.Sprintf("parent %s\n", parentSHA1))
	}

	// Timestamp and timezone
	timestamp := time.Now().Unix()
	_, offset := time.Now().Zone()
	timezone := fmt.Sprintf("%+03d%02d", offset/3600, (offset%3600)/60)

	// Author line
	authorLine := fmt.Sprintf("author %s <%s> %d %s\n",
		authorName, authorEmail, timestamp, timezone)
	commitContent.WriteString(authorLine)

	// Committer line (same as author for simplicity)
	committerLine := fmt.Sprintf("committer %s <%s> %d %s\n",
		authorName, authorEmail, timestamp, timezone)
	commitContent.WriteString(committerLine)

	// Empty line before message
	commitContent.WriteString("\n")

	// Commit message
	commitContent.WriteString(message)

	// Ensure message ends with newline
	if !strings.HasSuffix(message, "\n") {
		commitContent.WriteString("\n")
	}

	// Create commit object with header
	content := commitContent.Bytes()
	header := fmt.Sprintf("commit %d\x00", len(content))
	commitObj := append([]byte(header), content...)

	return commitObj, nil
}

// Helper: Get parent commit SHA-1 from current branch
func GetParentCommit(repoPath string) (string, error) {
	// Read HEAD to find current branch
	headPath := filepath.Join(repoPath, ".purr", "HEAD")
	headContent, err := os.ReadFile(headPath)
	if err != nil {
		return "", nil // No HEAD yet (first commit)
	}

	headStr := strings.TrimSpace(string(headContent))

	// Parse "ref: refs/heads/main"
	if strings.HasPrefix(headStr, "ref: ") {
		refPath := strings.TrimPrefix(headStr, "ref: ")
		branchPath := filepath.Join(repoPath, ".purr", refPath)

		parentSHA, err := os.ReadFile(branchPath)
		if err != nil {
			return "", nil // Branch exists but no commits yet
		}

		return strings.TrimSpace(string(parentSHA)), nil
	}

	// Detached HEAD case
	return headStr, nil
}

// Helper: Update branch reference with new commit
func UpdateBranchRef(repoPath, commitSHA1 string) error {
	// Read HEAD to find current branch
	headPath := filepath.Join(repoPath, ".purr", "HEAD")
	headContent, err := os.ReadFile(headPath)
	if err != nil {
		// No HEAD, create it pointing to main
		headContent = []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(headPath, headContent, 0644); err != nil {
			return fmt.Errorf("failed to create HEAD: %w", err)
		}
	}

	headStr := strings.TrimSpace(string(headContent))

	// Parse "ref: refs/heads/main"
	var branchPath string
	if strings.HasPrefix(headStr, "ref: ") {
		refPath := strings.TrimPrefix(headStr, "ref: ")
		branchPath = filepath.Join(repoPath, ".purr", refPath)
	} else {
		return fmt.Errorf("detached HEAD not supported yet")
	}

	// Create refs/heads directory if needed
	dir := filepath.Dir(branchPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create refs directory: %w", err)
	}

	// Write commit SHA-1 to branch file
	if err := os.WriteFile(branchPath, []byte(commitSHA1+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to update branch ref: %w", err)
	}

	return nil
}
