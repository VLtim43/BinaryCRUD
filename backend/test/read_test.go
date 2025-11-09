package test

import (
	"testing"

	"BinaryCRUD/backend/utils"
)

func TestReadFixedNumber(t *testing.T) {
	hexString := "0000002a" // 42 in 4 bytes
	value, newOffset, err := utils.ReadFixedNumber(4, hexString, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != 42 {
		t.Errorf("expected 42, got %d", value)
	}

	if newOffset != 8 {
		t.Errorf("expected offset 8, got %d", newOffset)
	}
}

func TestReadFixedNumberWithOffset(t *testing.T) {
	hexString := "ffff0005" // garbage + 5 in 2 bytes
	value, newOffset, err := utils.ReadFixedNumber(2, hexString, 4)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != 5 {
		t.Errorf("expected 5, got %d", value)
	}

	if newOffset != 8 {
		t.Errorf("expected offset 8, got %d", newOffset)
	}
}

func TestReadFixedNumberOverflow(t *testing.T) {
	hexString := "00ff"
	_, _, err := utils.ReadFixedNumber(4, hexString, 0)
	if err == nil {
		t.Error("expected error when not enough data, got none")
	}
}

func TestReadFixedNumberMaxValues(t *testing.T) {
	// Test 1 byte max (255)
	hexString := "ff"
	value, _, err := utils.ReadFixedNumber(1, hexString, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != 255 {
		t.Errorf("expected 255, got %d", value)
	}

	// Test 2 bytes max (65535)
	hexString = "ffff"
	value, _, err = utils.ReadFixedNumber(2, hexString, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != 65535 {
		t.Errorf("expected 65535, got %d", value)
	}
}
