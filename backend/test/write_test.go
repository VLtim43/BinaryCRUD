package test

import (
	"testing"

	"BinaryCRUD/backend/utils"
)

func TestWriteFixedString(t *testing.T) {
	result, err := utils.WriteFixedString(10, "hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "000000000068656c6c6f" // 5 zero bytes + "hello"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWriteVariable(t *testing.T) {
	result, err := utils.WriteVariable("hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "68656c6c6f" // "hello" in hex
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWriteFixedNumberTwoBytes(t *testing.T) {
	result, err := utils.WriteFixedNumber(2, 4)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "0004" // binary 4 in 2 bytes
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWriteFixedNumberTombstone(t *testing.T) {
	// Empty tombstone (0)
	result, err := utils.WriteFixedNumber(1, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "00" // binary 0
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	// Set tombstone (1)
	result, err = utils.WriteFixedNumber(1, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected = "01" // binary 1
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWriteHeader(t *testing.T) {
	// Test with simple values
	result, err := utils.WriteHeader(1, 2, 3)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expected format: [0000 0001][1F][0000 0002][1F][0000 0003]
	// 0x00000001 = "00000001"
	// 0x1F = "1f"
	// 0x00000002 = "00000002"
	// 0x00000003 = "00000003"
	expected := "000000011f000000021f00000003"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWriteHeaderWithZeros(t *testing.T) {
	// Test with zero values
	result, err := utils.WriteHeader(0, 0, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "000000001f000000001f00000000"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWriteHeaderWithLargeValues(t *testing.T) {
	// Test with larger values
	result, err := utils.WriteHeader(100, 50, 200)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 100 = 0x64, 50 = 0x32, 200 = 0xC8
	expected := "000000641f000000321f000000c8"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
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
	if result != "ff" {
		t.Errorf("expected ff, got %s", result)
	}

	result, err = utils.WriteFixedNumber(2, 65535)
	if err != nil {
		t.Errorf("unexpected error for max 2-byte value: %v", err)
	}
	if result != "ffff" {
		t.Errorf("expected ffff, got %s", result)
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
