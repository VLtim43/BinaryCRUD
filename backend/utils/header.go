package utils

import (
	"fmt"
)

// WriteHeader creates a header string with entitiesCount, tombstoneCount, and nextId
// separated by unit separators. Format: [count][0x1F][count][0x1F][count]
func WriteHeader(entitiesCount, tombstoneCount, nextId int) (string, error) {
	entitiesHex, err := WriteFixedNumber(4, uint64(entitiesCount))
	if err != nil {
		return "", fmt.Errorf("failed to write entitiesCount: %w", err)
	}

	tombstoneHex, err := WriteFixedNumber(4, uint64(tombstoneCount))
	if err != nil {
		return "", fmt.Errorf("failed to write tombstoneCount: %w", err)
	}

	nextIdHex, err := WriteFixedNumber(4, uint64(nextId))
	if err != nil {
		return "", fmt.Errorf("failed to write nextId: %w", err)
	}

	separatorHex, err := WriteVariable(UnitSeparator)
	if err != nil {
		return "", fmt.Errorf("failed to write separator: %w", err)
	}

	// Build the header: [entitiesCount][sep][tombstoneCount][sep][nextId]
	header := entitiesHex + separatorHex + tombstoneHex + separatorHex + nextIdHex

	return header, nil
}
