package persistence

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// OrderItem represents a single item in an order with quantity
type OrderItem struct {
	ItemID   uint32 `json:"itemId"`
	Quantity uint32 `json:"quantity"`
}

// Order represents an order with multiple items
type Order struct {
	RecordID  uint32
	Items     []OrderItem
	Tombstone bool
	Timestamp int64 // Unix timestamp in seconds
}

// AppendOrder writes an order to the binary file
func AppendOrder(filename string, items []OrderItem) (*AppendResult, error) {
	// Validate that we have items
	if len(items) == 0 {
		return nil, fmt.Errorf("cannot create empty order: items are required")
	}

	fmt.Printf("\n[DEBUG] === Appending Order with %d items ===\n", len(items))

	result, currentCount, err := appendBinaryRecord(
		filename,
		"order record",
		true,
		func(nextID uint32) (Order, error) {
			return Order{
				RecordID:  nextID,
				Items:     items,
				Tombstone: false,
				Timestamp: time.Now().Unix(),
			}, nil
		},
		WriteOrderRecord,
		func(count uint32, offset int64) {
			fmt.Printf("[DEBUG] Current order count: %d\n", count)
			fmt.Printf("[DEBUG] Writing order at offset: %d\n", offset)
		},
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] Updated header count to: %d\n", currentCount+1)
	fmt.Printf("[DEBUG] Assigned orderID: %d at offset: %d\n", result.RecordID, result.Offset)
	fmt.Printf("[DEBUG] === Order successfully written ===\n\n")

	return result, nil
}

// WriteOrderRecord writes a single order record to the writer.
// Format: [RecordID:4] [US:1] [Tombstone:1] [US:1] [ItemCount:4] [US:1] [Items...] [US:1] [Timestamp:8] [RS:1]
// Each item: [ItemID:4] [US:1] [Quantity:4] [US:1]
func WriteOrderRecord(writer *bufio.Writer, order Order, debug bool) error {
	itemCount := uint32(len(order.Items))

	// Write record ID (4 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, order.RecordID); err != nil {
		return fmt.Errorf("failed to write record ID: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote order record ID: [%02X %02X %02X %02X] (%d)\n",
			byte(order.RecordID), byte(order.RecordID>>8), byte(order.RecordID>>16), byte(order.RecordID>>24), order.RecordID)
	}

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write tombstone (0 = active, 1 = deleted)
	tombstone := uint8(0)
	if order.Tombstone {
		tombstone = 1
	}
	if err := binary.Write(writer, binary.LittleEndian, tombstone); err != nil {
		return fmt.Errorf("failed to write tombstone: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote tombstone: [%02X] (%s)\n", tombstone, getTombstoneStatus(order.Tombstone))
	}

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write item count (4 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, itemCount); err != nil {
		return fmt.Errorf("failed to write item count: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote item count: [%02X %02X %02X %02X] (%d items)\n",
			byte(itemCount), byte(itemCount>>8), byte(itemCount>>16), byte(itemCount>>24), itemCount)
	}

	// Write unit separator
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write each order item
	for idx, item := range order.Items {
		// Write ItemID (4 bytes, little-endian)
		if err := binary.Write(writer, binary.LittleEndian, item.ItemID); err != nil {
			return fmt.Errorf("failed to write item %d itemID: %w", idx, err)
		}
		if debug {
			fmt.Printf("[DEBUG] Wrote item %d ItemID: [%02X %02X %02X %02X] (%d)\n",
				idx, byte(item.ItemID), byte(item.ItemID>>8), byte(item.ItemID>>16), byte(item.ItemID>>24), item.ItemID)
		}

		// Write unit separator
		if err := writer.WriteByte(UnitSeparator); err != nil {
			return fmt.Errorf("failed to write unit separator: %w", err)
		}
		if debug {
			fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
		}

		// Write Quantity (4 bytes, little-endian)
		if err := binary.Write(writer, binary.LittleEndian, item.Quantity); err != nil {
			return fmt.Errorf("failed to write item %d quantity: %w", idx, err)
		}
		if debug {
			fmt.Printf("[DEBUG] Wrote item %d Quantity: [%02X %02X %02X %02X] (%d)\n",
				idx, byte(item.Quantity), byte(item.Quantity>>8), byte(item.Quantity>>16), byte(item.Quantity>>24), item.Quantity)
		}

		// Write unit separator
		if err := writer.WriteByte(UnitSeparator); err != nil {
			return fmt.Errorf("failed to write unit separator: %w", err)
		}
		if debug {
			fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
		}
	}

	// Write unit separator before timestamp
	if err := writer.WriteByte(UnitSeparator); err != nil {
		return fmt.Errorf("failed to write unit separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote unit separator: [%02X]\n", UnitSeparator)
	}

	// Write timestamp (8 bytes, little-endian)
	if err := binary.Write(writer, binary.LittleEndian, order.Timestamp); err != nil {
		return fmt.Errorf("failed to write timestamp: %w", err)
	}
	if debug {
		timestampBytes := make([]byte, 8)
		for i := 0; i < 8; i++ {
			timestampBytes[i] = byte(order.Timestamp >> (i * 8))
		}
		timestampDate := time.Unix(order.Timestamp, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("[DEBUG] Wrote timestamp: [%s] (%s)\n", formatHexBytesDebug(timestampBytes), timestampDate)
	}

	// Write record separator
	if err := writer.WriteByte(RecordSeparator); err != nil {
		return fmt.Errorf("failed to write record separator: %w", err)
	}
	if debug {
		fmt.Printf("[DEBUG] Wrote record separator: [%02X]\n", RecordSeparator)
	}

	return nil
}

// ReadOrderRecord reads a single order record from the reader
func ReadOrderRecord(reader *bufio.Reader) (*Order, error) {
	// Read record ID
	var recordID uint32
	if err := binary.Read(reader, binary.LittleEndian, &recordID); err != nil {
		return nil, fmt.Errorf("failed to read record ID: %w", err)
	}

	// Read and verify unit separator
	sep, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator after ID: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator after ID: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read tombstone
	var tombstone uint8
	if err := binary.Read(reader, binary.LittleEndian, &tombstone); err != nil {
		return nil, fmt.Errorf("failed to read tombstone: %w", err)
	}

	// Read and verify unit separator
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator after tombstone: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator after tombstone: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read item count (4 bytes, little-endian)
	var itemCount uint32
	if err := binary.Read(reader, binary.LittleEndian, &itemCount); err != nil {
		return nil, fmt.Errorf("failed to read item count: %w", err)
	}

	// Sanity check: item count shouldn't be larger than 10000 items (prevents corruption issues)
	if itemCount > 10000 {
		return nil, fmt.Errorf("invalid item count: %d (possibly corrupted file)", itemCount)
	}

	// Read and verify unit separator
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator after item count: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator after item count: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read all order items
	items := make([]OrderItem, itemCount)
	for i := uint32(0); i < itemCount; i++ {
		// Read ItemID (4 bytes, little-endian)
		var itemID uint32
		if err := binary.Read(reader, binary.LittleEndian, &itemID); err != nil {
			return nil, fmt.Errorf("failed to read item %d itemID: %w", i, err)
		}

		// Read and verify unit separator
		sep, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read unit separator after item %d itemID: %w", i, err)
		}
		if sep != UnitSeparator {
			return nil, fmt.Errorf("invalid unit separator after item %d itemID: expected 0x%02X, got 0x%02X", i, UnitSeparator, sep)
		}

		// Read Quantity (4 bytes, little-endian)
		var quantity uint32
		if err := binary.Read(reader, binary.LittleEndian, &quantity); err != nil {
			return nil, fmt.Errorf("failed to read item %d quantity: %w", i, err)
		}

		// Read and verify unit separator
		sep, err = reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read unit separator after item %d quantity: %w", i, err)
		}
		if sep != UnitSeparator {
			return nil, fmt.Errorf("invalid unit separator after item %d quantity: expected 0x%02X, got 0x%02X", i, UnitSeparator, sep)
		}

		items[i] = OrderItem{
			ItemID:   itemID,
			Quantity: quantity,
		}
	}

	// Read and verify unit separator before timestamp
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read unit separator before timestamp: %w", err)
	}
	if sep != UnitSeparator {
		return nil, fmt.Errorf("invalid unit separator before timestamp: expected 0x%02X, got 0x%02X", UnitSeparator, sep)
	}

	// Read timestamp (8 bytes, little-endian)
	var timestamp int64
	if err := binary.Read(reader, binary.LittleEndian, &timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}

	// Read and verify record separator
	sep, err = reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read record separator: %w", err)
	}
	if sep != RecordSeparator {
		return nil, fmt.Errorf("invalid record separator: expected 0x%02X, got 0x%02X", RecordSeparator, sep)
	}

	return &Order{
		RecordID:  recordID,
		Items:     items,
		Tombstone: tombstone != 0,
		Timestamp: timestamp,
	}, nil
}

// ReadAllOrders reads all orders from the binary file
func ReadAllOrders(filename string) ([]Order, error) {
	return readRecords(
		filename,
		"order record",
		ReadOrderRecord,
		nil,
		func(order *Order) bool {
			return !order.Tombstone
		},
	)
}

// PrintOrderBinaryFile prints the orders binary file in a human-readable format
func PrintOrderBinaryFile(filename string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	output := fmt.Sprintf("filename: %s\n", filename)

	reader := bufio.NewReader(file)

	// Read header
	count, countBytes, err := readHeaderForPrint(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read header: %w", err)
	}
	output += fmt.Sprintf("order count: %d [%s]\n", count, formatHexBytes(countBytes))
	output += "-------------------------\n"

	// Read and print all orders
	for i := uint32(0); i < count; i++ {
		order, err := ReadOrderRecord(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read order %d: %w", i, err)
		}

		status := "active"
		if order.Tombstone {
			status = "deleted"
		}

		timestampDate := time.Unix(order.Timestamp, 0).Format("2006-01-02 15:04:05")

		output += fmt.Sprintf("order id: %d\n", order.RecordID)
		output += fmt.Sprintf("timestamp: %s\n", timestampDate)
		output += fmt.Sprintf("status: %s\n", status)
		output += fmt.Sprintf("items (%d):\n", len(order.Items))
		for _, item := range order.Items {
			output += fmt.Sprintf("  - ItemID: %d, Quantity: %d\n", item.ItemID, item.Quantity)
		}
		output += "-------------------------\n"
	}

	return output, nil
}
