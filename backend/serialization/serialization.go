package serialization

import (
	"os"
	"path/filepath"
)

func WriteHex(filename string, text string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(text), 0644)
}

func ReadHex(filename string) (string, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
