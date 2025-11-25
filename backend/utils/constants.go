package utils

// BDATMagic is the magic bytes for binary data files
var BDATMagic = []byte{'B', 'D', 'A', 'T'}

const (
	// IDSize is the size of the ID field in bytes
	IDSize = 2

	// TombstoneSize is the size of the tombstone field in bytes
	TombstoneSize = 1

	// RecordLengthSize is the size of the record length prefix in bytes
	RecordLengthSize = 2

	// HeaderFieldSize is the size of each header field in bytes
	HeaderFieldSize = 4

	// MagicSize is the size of the magic bytes
	MagicSize = 4

	// FilenameLengthSize is the size of the filename length field
	FilenameLengthSize = 1

	// HeaderFixedSize is the fixed portion of the header (magic + counts)
	// Format: [magic(4)][filenameLen(1)][filename(N)][entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	// The variable part is filename, fixed part = 4 + 1 + 4 + 4 + 4 = 17 bytes + filename
	HeaderFixedSize = MagicSize + FilenameLengthSize + (HeaderFieldSize * 3)

	// DefaultBTreeOrder is the default order for B+ tree indices
	DefaultBTreeOrder = 4
)

// CalculateHeaderSize returns the total header size for a given filename
func CalculateHeaderSize(filename string) int {
	return HeaderFixedSize + len(filename)
}
