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
func CleanupDataFiles(log LogFunc) ([]FolderCleanupResult, error) {
	foldersToClean := []string{
		BinDir,
		IndexDir,
		CompressedDir,
	}

	results := make([]FolderCleanupResult, 0, len(foldersToClean))

	for _, folder := range foldersToClean {
		count, err := cleanFolder(folder, log)
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
	idxPattern := filepath.Join(IndexDir, prefix+"*")
	idxMatches, _ := filepath.Glob(idxPattern)
	for _, match := range idxMatches {
		os.Remove(match)
	}

	return nil
}

// cleanFolder removes all files (not directories) from a folder
func cleanFolder(folder string, log LogFunc) (int, error) {
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
		if err := RemoveFile(filePath, log); err == nil {
			count++
		}
	}

	return count, nil
}

// CleanupTempFiles removes leftover temp files (.tmp) from index directory
func CleanupTempFiles() error {
	if _, err := os.Stat(IndexDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(IndexDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".tmp" {
			os.Remove(filepath.Join(IndexDir, entry.Name()))
		}
	}

	return nil
}
