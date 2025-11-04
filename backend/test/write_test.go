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
