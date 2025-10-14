package serialization

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
)

func AppendEntry(filename string, name string) error {
	if err := InitFile(filename); err != nil {
		return err
	}

	fmt.Printf("\n[DEBUG] === Appending Entry: \"%s\" ===\n", name)

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var count uint32
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return err
	}

	fmt.Printf("[DEBUG] Current record count: %d\n", count)

	if _, err := file.Seek(0, 2); err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	// Write tombstone (0 = active)
	if err := binary.Write(writer, binary.LittleEndian, uint8(0)); err != nil {
		return err
	}
	fmt.Printf("[DEBUG] Wrote tombstone: [00] (active)\n")

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return err
	}
	fmt.Printf("[DEBUG] Wrote unit separator: [1F]\n")

	// Write name length
	size := uint32(len(name))
	if err := binary.Write(writer, binary.LittleEndian, size); err != nil {
		return err
	}
	fmt.Printf("[DEBUG] Wrote name length: [%02X %02X %02X %02X] (%d bytes)\n",
		byte(size), byte(size>>8), byte(size>>16), byte(size>>24), size)

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return err
	}
	fmt.Printf("[DEBUG] Wrote unit separator: [1F]\n")

	// Write name data
	nameBytes := []byte(name)
	if _, err := writer.Write(nameBytes); err != nil {
		return err
	}
	fmt.Printf("[DEBUG] Wrote name data: [%s] (\"%s\")\n", hex.EncodeToString(nameBytes), name)

	// Write record separator
	if err := writer.WriteByte(RecordSeparator); err != nil {
		return err
	}
	fmt.Printf("[DEBUG] Wrote record separator: [1E]\n")

	if err := writer.Flush(); err != nil {
		return err
	}

	// Seek back to start to update header
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// Update record count in header
	writer = bufio.NewWriter(file)
	count++
	if err := binary.Write(writer, binary.LittleEndian, count); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	fmt.Printf("[DEBUG] Updated header count to: %d\n", count)
	fmt.Printf("[DEBUG] === Entry successfully written ===\n\n")

	return nil
}
