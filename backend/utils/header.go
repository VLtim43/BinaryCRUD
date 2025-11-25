package utils

import (
	"bytes"
	"fmt"
	"os"
)

// WriteHeader creates a header byte slice with filename and counts
// Format: [magic(4)][filenameLen(1)][filename(N)][entitiesCount(4)][tombstoneCount(4)][nextId(4)]
func WriteHeader(filename string, entitiesCount, tombstoneCount, nextId int) ([]byte, error) {
	if len(filename) > 255 {
		return nil, fmt.Errorf("filename too long: max 255 bytes, got %d", len(filename))
	}

	var header bytes.Buffer

	// Magic bytes
	header.Write(BDATMagic)

	// Filename length (1 byte)
	header.WriteByte(byte(len(filename)))

	// Filename
	header.WriteString(filename)

	// Entities count (4 bytes)
	entitiesBytes, err := WriteFixedNumber(HeaderFieldSize, uint64(entitiesCount))
	if err != nil {
		return nil, fmt.Errorf("failed to write entitiesCount: %w", err)
	}
	header.Write(entitiesBytes)

	// Tombstone count (4 bytes)
	tombstoneBytes, err := WriteFixedNumber(HeaderFieldSize, uint64(tombstoneCount))
	if err != nil {
		return nil, fmt.Errorf("failed to write tombstoneCount: %w", err)
	}
	header.Write(tombstoneBytes)

	// Next ID (4 bytes)
	nextIdBytes, err := WriteFixedNumber(HeaderFieldSize, uint64(nextId))
	if err != nil {
		return nil, fmt.Errorf("failed to write nextId: %w", err)
	}
	header.Write(nextIdBytes)

	return header.Bytes(), nil
}

// ReadHeader reads and parses the header from a file
// Returns (filename, entitiesCount, tombstoneCount, nextId, error)
func ReadHeader(file *os.File) (string, int, int, int, error) {
	// Seek to beginning
	_, err := file.Seek(0, 0)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Read magic bytes
	magic := make([]byte, MagicSize)
	n, err := file.Read(magic)
	if err != nil || n != MagicSize {
		return "", 0, 0, 0, fmt.Errorf("failed to read magic bytes")
	}
	if !bytes.Equal(magic, BDATMagic) {
		return "", 0, 0, 0, fmt.Errorf("invalid magic bytes: expected BDAT")
	}

	// Read filename length
	filenameLenByte := make([]byte, 1)
	n, err = file.Read(filenameLenByte)
	if err != nil || n != 1 {
		return "", 0, 0, 0, fmt.Errorf("failed to read filename length")
	}
	filenameLen := int(filenameLenByte[0])

	// Read filename
	filenameBytes := make([]byte, filenameLen)
	n, err = file.Read(filenameBytes)
	if err != nil || n != filenameLen {
		return "", 0, 0, 0, fmt.Errorf("failed to read filename")
	}
	filename := string(filenameBytes)

	// Read counts (3 x 4 bytes)
	countsBytes := make([]byte, HeaderFieldSize*3)
	n, err = file.Read(countsBytes)
	if err != nil || n != HeaderFieldSize*3 {
		return "", 0, 0, 0, fmt.Errorf("failed to read counts")
	}

	offset := 0
	entitiesCount, offset, err := ReadFixedNumber(HeaderFieldSize, countsBytes, offset)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to read entitiesCount: %w", err)
	}

	tombstoneCount, offset, err := ReadFixedNumber(HeaderFieldSize, countsBytes, offset)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to read tombstoneCount: %w", err)
	}

	nextId, _, err := ReadFixedNumber(HeaderFieldSize, countsBytes, offset)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to read nextId: %w", err)
	}

	return filename, int(entitiesCount), int(tombstoneCount), int(nextId), nil
}

// ReadHeaderFromBytes reads header from a byte slice (for archive unpacking)
// Returns (filename, entitiesCount, tombstoneCount, nextId, headerSize, error)
func ReadHeaderFromBytes(data []byte) (string, int, int, int, int, error) {
	if len(data) < MagicSize+FilenameLengthSize {
		return "", 0, 0, 0, 0, fmt.Errorf("data too short for header")
	}

	// Check magic
	if !bytes.Equal(data[:MagicSize], BDATMagic) {
		return "", 0, 0, 0, 0, fmt.Errorf("invalid magic bytes: expected BDAT")
	}

	// Read filename length
	filenameLen := int(data[MagicSize])
	headerSize := CalculateHeaderSize(string(data[MagicSize+FilenameLengthSize : MagicSize+FilenameLengthSize+filenameLen]))

	if len(data) < headerSize {
		return "", 0, 0, 0, 0, fmt.Errorf("data too short for header with filename")
	}

	// Read filename
	filename := string(data[MagicSize+FilenameLengthSize : MagicSize+FilenameLengthSize+filenameLen])

	// Read counts
	countsStart := MagicSize + FilenameLengthSize + filenameLen
	countsBytes := data[countsStart : countsStart+HeaderFieldSize*3]

	offset := 0
	entitiesCount, offset, err := ReadFixedNumber(HeaderFieldSize, countsBytes, offset)
	if err != nil {
		return "", 0, 0, 0, 0, fmt.Errorf("failed to read entitiesCount: %w", err)
	}

	tombstoneCount, offset, err := ReadFixedNumber(HeaderFieldSize, countsBytes, offset)
	if err != nil {
		return "", 0, 0, 0, 0, fmt.Errorf("failed to read tombstoneCount: %w", err)
	}

	nextId, _, err := ReadFixedNumber(HeaderFieldSize, countsBytes, offset)
	if err != nil {
		return "", 0, 0, 0, 0, fmt.Errorf("failed to read nextId: %w", err)
	}

	return filename, int(entitiesCount), int(tombstoneCount), int(nextId), headerSize, nil
}

// UpdateHeader updates the header in the file with new values (keeps same filename)
func UpdateHeader(file *os.File, entitiesCount, tombstoneCount, nextId int) error {
	// First read current filename
	filename, _, _, _, err := ReadHeader(file)
	if err != nil {
		return fmt.Errorf("failed to read current header: %w", err)
	}

	// Generate new header with same filename
	header, err := WriteHeader(filename, entitiesCount, tombstoneCount, nextId)
	if err != nil {
		return fmt.Errorf("failed to generate header: %w", err)
	}

	// Seek to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Write the header
	_, err = file.Write(header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Force write to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync header to disk: %w", err)
	}

	return nil
}

// WriteHeaderToFile writes a header to the beginning of an empty file
func WriteHeaderToFile(file *os.File, header []byte) error {
	// Check if file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() != 0 {
		return fmt.Errorf("file is not empty, cannot write header")
	}

	// Seek to the beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to beginning: %w", err)
	}

	// Write the header
	err = WriteToFile(file, header)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Force write to disk
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync header to disk: %w", err)
	}

	return nil
}
