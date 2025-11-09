package test

import (
	"testing"

	"BinaryCRUD/backend/utils"
)

func TestReadFixedNumber(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x2a} // 42 in 4 bytes
	value, newOffset, err := utils.ReadFixedNumber(4, data, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != 42 {
		t.Errorf("expected 42, got %d", value)
	}

	if newOffset != 4 {
		t.Errorf("expected offset 4, got %d", newOffset)
	}
}

func TestReadFixedNumberWithOffset(t *testing.T) {
	data := []byte{0xff, 0xff, 0x00, 0x05} // garbage + 5 in 2 bytes
	value, newOffset, err := utils.ReadFixedNumber(2, data, 2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != 5 {
		t.Errorf("expected 5, got %d", value)
	}

	if newOffset != 4 {
		t.Errorf("expected offset 4, got %d", newOffset)
	}
}

func TestReadFixedNumberOverflow(t *testing.T) {
	data := []byte{0x00, 0xff}
	_, _, err := utils.ReadFixedNumber(4, data, 0)
	if err == nil {
		t.Error("expected error when not enough data, got none")
	}
}

func TestReadFixedNumberMaxValues(t *testing.T) {
	// Test 1 byte max (255)
	data := []byte{0xff}
	value, _, err := utils.ReadFixedNumber(1, data, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != 255 {
		t.Errorf("expected 255, got %d", value)
	}

	// Test 2 bytes max (65535)
	data = []byte{0xff, 0xff}
	value, _, err = utils.ReadFixedNumber(2, data, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != 65535 {
		t.Errorf("expected 65535, got %d", value)
	}
}
