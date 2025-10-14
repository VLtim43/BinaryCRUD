package serialization

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type Item struct {
	Name      string
	Tombstone bool
}

const (
	HeaderSize        = 4
	RecordSeparator   = 0x1E
	UnitSeparator     = 0x1F
)

func InitFile(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(filename); err == nil {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	if err := binary.Write(writer, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}

	if err := writer.WriteByte(RecordSeparator); err != nil {
		return err
	}

	return writer.Flush()
}

func AppendEntry(filename string, name string) error {
	if err := InitFile(filename); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var count uint32
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return err
	}

	if _, err := file.Seek(0, 2); err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	if err := binary.Write(writer, binary.LittleEndian, uint8(0)); err != nil {
		return err
	}

	if err := writer.WriteByte(UnitSeparator); err != nil {
		return err
	}

	size := uint32(len(name))
	if err := binary.Write(writer, binary.LittleEndian, size); err != nil {
		return err
	}

	if err := writer.WriteByte(UnitSeparator); err != nil {
		return err
	}

	if _, err := writer.Write([]byte(name)); err != nil {
		return err
	}

	if err := writer.WriteByte(RecordSeparator); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	writer = bufio.NewWriter(file)
	count++
	if err := binary.Write(writer, binary.LittleEndian, count); err != nil {
		return err
	}

	return writer.Flush()
}

func ReadAllEntries(filename string) ([]Item, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []Item{}, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var count uint32
	if err := binary.Read(reader, binary.LittleEndian, &count); err != nil {
		return nil, err
	}

	if _, err := reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read record separator: %w", err)
	}

	items := make([]Item, 0, count)

	for i := uint32(0); i < count; i++ {
		var tombstone uint8
		if err := binary.Read(reader, binary.LittleEndian, &tombstone); err != nil {
			return nil, fmt.Errorf("failed to read tombstone: %w", err)
		}

		if _, err := reader.ReadByte(); err != nil {
			return nil, fmt.Errorf("failed to read unit separator: %w", err)
		}

		var size uint32
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			return nil, fmt.Errorf("failed to read size: %w", err)
		}

		if _, err := reader.ReadByte(); err != nil {
			return nil, fmt.Errorf("failed to read unit separator: %w", err)
		}

		data := make([]byte, size)
		if _, err := reader.Read(data); err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}

		if _, err := reader.ReadByte(); err != nil {
			return nil, fmt.Errorf("failed to read record separator: %w", err)
		}

		if tombstone == 0 {
			items = append(items, Item{
				Name:      string(data),
				Tombstone: false,
			})
		}
	}

	return items, nil
}
