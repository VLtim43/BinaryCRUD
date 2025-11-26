package test

import (
	"bytes"
	"testing"

	"BinaryCRUD/backend/utils"
)

func TestWriteFixedString(t *testing.T) {
	result, err := utils.WriteFixedString(10, "hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x68, 0x65, 0x6c, 0x6c, 0x6f} // 5 zero bytes + "hello"
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteVariable(t *testing.T) {
	result, err := utils.WriteVariable("hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []byte("hello")
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteFixedNumberTwoBytes(t *testing.T) {
	result, err := utils.WriteFixedNumber(2, 4)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []byte{0x00, 0x04} // binary 4 in 2 bytes
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteFixedNumberTombstone(t *testing.T) {
	// Empty tombstone (0)
	result, err := utils.WriteFixedNumber(1, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []byte{0x00} // binary 0
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Set tombstone (1)
	result, err = utils.WriteFixedNumber(1, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected = []byte{0x01} // binary 1
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteHeader(t *testing.T) {
	// Test with simple values
	result, err := utils.WriteHeader("test.bin", 1, 2, 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expected format: [magic(4)][filenameLen(1)][filename(N)][entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	// For "test.bin" (8 bytes): 4 + 1 + 8 + 4 + 4 + 4 = 25 bytes
	expected := []byte{
		'B', 'D', 'A', 'T', // magic
		8,                            // filename length
		't', 'e', 's', 't', '.', 'b', 'i', 'n', // filename
		0x00, 0x00, 0x00, 0x01, // entitiesCount = 1
		0x00, 0x00, 0x00, 0x02, // tombstoneCount = 2
		0x00, 0x00, 0x00, 0x03, // nextId = 3
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteHeaderWithZeros(t *testing.T) {
	// Test with zero values (empty filename)
	result, err := utils.WriteHeader("", 0, 0, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expected format: [magic(4)][filenameLen(1)][filename(0)][entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	// For empty filename: 4 + 1 + 0 + 4 + 4 + 4 = 17 bytes
	expected := []byte{
		'B', 'D', 'A', 'T', // magic
		0,                  // filename length = 0
		0x00, 0x00, 0x00, 0x00, // entitiesCount = 0
		0x00, 0x00, 0x00, 0x00, // tombstoneCount = 0
		0x00, 0x00, 0x00, 0x00, // nextId = 0
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteHeaderWithLargeValues(t *testing.T) {
	// Test with larger values
	result, err := utils.WriteHeader("data.bin", 100, 50, 200)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expected format: [magic(4)][filenameLen(1)][filename(N)][entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	// 100 = 0x64, 50 = 0x32, 200 = 0xC8
	// For "data.bin" (8 bytes): 4 + 1 + 8 + 4 + 4 + 4 = 25 bytes
	expected := []byte{
		'B', 'D', 'A', 'T', // magic
		8,                            // filename length
		'd', 'a', 't', 'a', '.', 'b', 'i', 'n', // filename
		0x00, 0x00, 0x00, 0x64, // entitiesCount = 100
		0x00, 0x00, 0x00, 0x32, // tombstoneCount = 50
		0x00, 0x00, 0x00, 0xc8, // nextId = 200
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteFixedNumberOverflow(t *testing.T) {
	// Test 1 byte overflow (max is 255)
	_, err := utils.WriteFixedNumber(1, 256)
	if err == nil {
		t.Error("expected error when value exceeds 1 byte capacity, got none")
	}

	// Test 2 byte overflow (max is 65535)
	_, err = utils.WriteFixedNumber(2, 65536)
	if err == nil {
		t.Error("expected error when value exceeds 2 byte capacity, got none")
	}

	// Test valid max values
	result, err := utils.WriteFixedNumber(1, 255)
	if err != nil {
		t.Errorf("unexpected error for max 1-byte value: %v", err)
	}
	expected := []byte{0xff}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	result, err = utils.WriteFixedNumber(2, 65535)
	if err != nil {
		t.Errorf("unexpected error for max 2-byte value: %v", err)
	}
	expected = []byte{0xff, 0xff}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteFixedStringOverflow(t *testing.T) {
	// Test string too long for size
	_, err := utils.WriteFixedString(3, "hello")
	if err == nil {
		t.Error("expected error when string exceeds size, got none")
	}

	_, err = utils.WriteFixedString(1, "ab")
	if err == nil {
		t.Error("expected error when string exceeds size, got none")
	}
}
