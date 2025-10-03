package serialization

import (
	"os"
)

func WriteHex(filename string, text string) error {
	return os.WriteFile(filename, []byte(text), 0644)
}

func ReadHex(filename string) (string, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
