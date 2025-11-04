package test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"BinaryCRUD/backend/utils"
)

func TestWriteFixed(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		content     []byte
		expectError bool
		validate    func(t *testing.T, result []byte)
	}{
		{
			name:        "content shorter than size - should pad with zeros",
			size:        10,
			content:     []byte("hello"),
			expectError: false,
			validate: func(t *testing.T, result []byte) {
				if len(result) != 10 {
					t.Errorf("expected length 10, got %d", len(result))
				}
				expected := []byte{'h', 'e', 'l', 'l', 'o', 0, 0, 0, 0, 0}
				if !bytes.Equal(result, expected) {
					t.Errorf("expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:        "content equal to size",
			size:        5,
			content:     []byte("hello"),
			expectError: false,
			validate: func(t *testing.T, result []byte) {
				if len(result) != 5 {
					t.Errorf("expected length 5, got %d", len(result))
				}
				if !bytes.Equal(result, []byte("hello")) {
					t.Errorf("expected %v, got %v", []byte("hello"), result)
				}
			},
		},
		{
			name:        "content longer than size - should truncate",
			size:        3,
			content:     []byte("hello"),
			expectError: false,
			validate: func(t *testing.T, result []byte) {
				if len(result) != 3 {
					t.Errorf("expected length 3, got %d", len(result))
				}
				if !bytes.Equal(result, []byte("hel")) {
					t.Errorf("expected %v, got %v", []byte("hel"), result)
				}
			},
		},
		{
			name:        "empty content",
			size:        5,
			content:     []byte{},
			expectError: false,
			validate: func(t *testing.T, result []byte) {
				if len(result) != 5 {
					t.Errorf("expected length 5, got %d", len(result))
				}
				expected := []byte{0, 0, 0, 0, 0}
				if !bytes.Equal(result, expected) {
					t.Errorf("expected %v, got %v", expected, result)
				}
			},
		},
		{
			name:        "zero size - should error",
			size:        0,
			content:     []byte("hello"),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "negative size - should error",
			size:        -5,
			content:     []byte("hello"),
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.WriteFixed(tt.size, tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestWriteVariable(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		validate func(t *testing.T, result []byte)
	}{
		{
			name:    "normal content",
			content: []byte("hello"),
			validate: func(t *testing.T, result []byte) {
				if len(result) != 9 {
					t.Errorf("expected length 9, got %d", len(result))
				}

				// Read the size from first 4 bytes
				var size uint32
				buf := bytes.NewReader(result[:4])
				if err := binary.Read(buf, binary.LittleEndian, &size); err != nil {
					t.Errorf("error reading size: %v", err)
				}

				if size != 5 {
					t.Errorf("expected size 5, got %d", size)
				}

				if !bytes.Equal(result[4:], []byte("hello")) {
					t.Errorf("expected content 'hello', got %v", result[4:])
				}
			},
		},
		{
			name:    "empty content",
			content: []byte{},
			validate: func(t *testing.T, result []byte) {
				if len(result) != 4 {
					t.Errorf("expected length 4, got %d", len(result))
				}

				var size uint32
				buf := bytes.NewReader(result[:4])
				if err := binary.Read(buf, binary.LittleEndian, &size); err != nil {
					t.Errorf("error reading size: %v", err)
				}

				if size != 0 {
					t.Errorf("expected size 0, got %d", size)
				}
			},
		},
		{
			name:    "large content",
			content: bytes.Repeat([]byte("A"), 1000),
			validate: func(t *testing.T, result []byte) {
				if len(result) != 1004 {
					t.Errorf("expected length 1004, got %d", len(result))
				}

				var size uint32
				buf := bytes.NewReader(result[:4])
				if err := binary.Read(buf, binary.LittleEndian, &size); err != nil {
					t.Errorf("error reading size: %v", err)
				}

				if size != 1000 {
					t.Errorf("expected size 1000, got %d", size)
				}

				if !bytes.Equal(result[4:], bytes.Repeat([]byte("A"), 1000)) {
					t.Error("content mismatch")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.WriteVariable(tt.content)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
