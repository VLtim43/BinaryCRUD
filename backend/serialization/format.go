package serialization

// BinaryFormat defines the structure and layout of the binary file format.
// This is the single source of truth for all binary format specifications.
//
// File Structure:
//   HEADER: [RecordCount:4bytes] [UnitSeparator:1byte] [NextID:4bytes] [RecordSeparator:1byte]
//   RECORD: [ID:4bytes] [UnitSeparator:1byte] [Tombstone:1byte] [UnitSeparator:1byte] [NameSize:4bytes] [UnitSeparator:1byte] [NameData:Nbytes] [UnitSeparator:1byte] [Timestamp:8bytes] [RecordSeparator:1byte]
//
// Example:
//   Header: [03 00 00 00 1F 03 00 00 00 1E] = 3 records, nextID=3
//   Record: [00 00 00 00 1F 00 1F 05 00 00 00 1F 70 69 7A 7A 61 1F <timestamp> 1E] = ID:0 "pizza" (active)
type BinaryFormat struct {
	// Header structure
	HeaderCountSize     int // 4 bytes - uint32 record count
	HeaderNextIDSize    int // 4 bytes - uint32 next ID counter
	HeaderSeparatorSize int // 1 byte - record separator

	// Record structure
	IDSize              int // 4 bytes - uint32 record ID
	TombstoneSize       int // 1 byte - uint8 (0=active, 1=deleted)
	UnitSeparatorSize   int // 1 byte - unit separator
	NameSizeFieldSize   int // 4 bytes - uint32 length of name
	TimestampSize       int // 8 bytes - int64 unix timestamp
	RecordSeparatorSize int // 1 byte - record separator
}

// GetFormat returns the current binary format specification
func GetFormat() *BinaryFormat {
	return &BinaryFormat{
		HeaderCountSize:     4,
		HeaderNextIDSize:    4,
		HeaderSeparatorSize: 1,
		IDSize:              4,
		TombstoneSize:       1,
		UnitSeparatorSize:   1,
		NameSizeFieldSize:   4,
		TimestampSize:       8,
		RecordSeparatorSize: 1,
	}
}

// HeaderSize returns the total size of the header in bytes
func (f *BinaryFormat) HeaderSize() int {
	return f.HeaderCountSize + f.UnitSeparatorSize + f.HeaderNextIDSize + f.HeaderSeparatorSize
}

// RecordOverheadSize returns the size of non-data bytes in a record
func (f *BinaryFormat) RecordOverheadSize() int {
	return f.IDSize +
		f.UnitSeparatorSize + // After ID
		f.TombstoneSize +
		f.UnitSeparatorSize + // After tombstone
		f.NameSizeFieldSize +
		f.UnitSeparatorSize + // After name size
		f.UnitSeparatorSize + // Before timestamp
		f.TimestampSize +
		f.RecordSeparatorSize
}

// CalculateRecordSize returns the total size of a record including data
func (f *BinaryFormat) CalculateRecordSize(nameLength int) int {
	return f.RecordOverheadSize() + nameLength
}

// MinimumRecordSize returns the smallest possible record size (empty string)
func (f *BinaryFormat) MinimumRecordSize() int {
	return f.RecordOverheadSize()
}
