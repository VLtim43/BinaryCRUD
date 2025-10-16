package serialization

type BinaryFormat struct {
	// Header structure
	HeaderCountSize     int // 4 bytes - uint32 record count
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
	return f.HeaderCountSize + f.HeaderSeparatorSize
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
