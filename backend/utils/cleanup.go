package utils

import (
	"os"
	"path/filepath"
)

// FolderCleanupResult contains the result of cleaning a single folder
type FolderCleanupResult struct {
	Folder string
	Count  int
}

// CleanupDataFiles deletes all generated data files (bin, indexes, compressed)
// but preserves seed data. Returns per-folder results.
func CleanupDataFiles() ([]FolderCleanupResult, error) {
	foldersToClean := []string{
		filepath.Join("data", "bin"),
		filepath.Join("data", "indexes"),
		filepath.Join("data", "compressed"),
	}

	results := make([]FolderCleanupResult, 0, len(foldersToClean))

	for _, folder := range foldersToClean {
		count, err := cleanFolder(folder)
		if err != nil {
			return results, err
		}
		results = append(results, FolderCleanupResult{
			Folder: folder,
			Count:  count,
		})
	}

	return results, nil
}

// CleanupTestFiles deletes test files from /tmp with a given prefix
func CleanupTestFiles(prefix string) error {
	pattern := filepath.Join("/tmp", prefix+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, match := range matches {
		os.Remove(match)
	}

	// Also clean up any index files in data/indexes that match test patterns
	idxPattern := filepath.Join("data", "indexes", prefix+"*")
	idxMatches, _ := filepath.Glob(idxPattern)
	for _, match := range idxMatches {
		os.Remove(match)
	}

	return nil
}

// cleanFolder removes all files (not directories) from a folder
func cleanFolder(folder string) (int, error) {
	// Check if folder exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return 0, nil
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(folder, entry.Name())
		if err := os.Remove(filePath); err == nil {
			count++
		}
	}

	return count, nil
}

// CleanupTempFiles removes leftover temp files (.tmp) from index directory
func CleanupTempFiles() error {
	indexDir := filepath.Join("data", "indexes")

	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(indexDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".tmp" {
			os.Remove(filepath.Join(indexDir, entry.Name()))
		}
	}

	return nil
}
