package serialization

// BinaryFormat defines the structure and layout of the binary file format.
// This is the single source of truth for all binary format specifications.
//
// File Structure:
//   HEADER: [RecordCount:4bytes] [RecordSeparator:1byte]
//   RECORD: [Tombstone:1byte] [UnitSeparator:1byte] [NameSize:4bytes] [UnitSeparator:1byte] [NameData:Nbytes] [RecordSeparator:1byte]
//
// Example:
//   Header: [03 00 00 00 1E] = 3 records
//   Record: [00 1F 05 00 00 00 1F 70 69 7A 7A 61 1E] = "pizza" (active)
type BinaryFormat struct {
	// Header structure
	HeaderCountSize     int // 4 bytes - uint32 record count
	HeaderSeparatorSize int // 1 byte - record separator

	// Record structure
	TombstoneSize      int // 1 byte - uint8 (0=active, 1=deleted)
	UnitSeparatorSize  int // 1 byte - unit separator
	NameSizeFieldSize  int // 4 bytes - uint32 length of name
	RecordSeparatorSize int // 1 byte - record separator
}

// GetFormat returns the current binary format specification
func GetFormat() *BinaryFormat {
	return &BinaryFormat{
		HeaderCountSize:     4,
		HeaderSeparatorSize: 1,
		TombstoneSize:       1,
		UnitSeparatorSize:   1,
		NameSizeFieldSize:   4,
		RecordSeparatorSize: 1,
	}
}

// HeaderSize returns the total size of the header in bytes
func (f *BinaryFormat) HeaderSize() int {
	return f.HeaderCountSize + f.HeaderSeparatorSize
}

// RecordOverheadSize returns the size of non-data bytes in a record
func (f *BinaryFormat) RecordOverheadSize() int {
	return f.TombstoneSize +
		f.UnitSeparatorSize +
		f.NameSizeFieldSize +
		f.UnitSeparatorSize +
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
