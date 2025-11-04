package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func WriteFixed(size int, content []byte) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}

	buf := new(bytes.Buffer)

	data := make([]byte, size)

	copy(data, content)

	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func WriteVariable(content []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	size := uint32(len(content))
	if err := binary.Write(buf, binary.LittleEndian, size); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, content); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
