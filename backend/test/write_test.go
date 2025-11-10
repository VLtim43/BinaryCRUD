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
	result, err := utils.WriteHeader(1, 2, 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expected format: [0000 0001][1F][0000 0002][1F][0000 0003][1E]
	expected := []byte{0x00, 0x00, 0x00, 0x01, 0x1f, 0x00, 0x00, 0x00, 0x02, 0x1f, 0x00, 0x00, 0x00, 0x03, 0x1e}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteHeaderWithZeros(t *testing.T) {
	// Test with zero values
	result, err := utils.WriteHeader(0, 0, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []byte{0x00, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x1e}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestWriteHeaderWithLargeValues(t *testing.T) {
	// Test with larger values
	result, err := utils.WriteHeader(100, 50, 200)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 100 = 0x64, 50 = 0x32, 200 = 0xC8
	expected := []byte{0x00, 0x00, 0x00, 0x64, 0x1f, 0x00, 0x00, 0x00, 0x32, 0x1f, 0x00, 0x00, 0x00, 0xc8, 0x1e}
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
