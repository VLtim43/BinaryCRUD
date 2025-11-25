package utils

const (
	// IDSize is the size of the ID field in bytes
	IDSize = 2

	// TombstoneSize is the size of the tombstone field in bytes
	TombstoneSize = 1

	// RecordLengthSize is the size of the record length prefix in bytes
	RecordLengthSize = 2

	// HeaderFieldSize is the size of each header field in bytes
	HeaderFieldSize = 4

	// FileNameSize is the size of the file name field in the header (fixed size)
	FileNameSize = 32

	// HeaderSize is the total size of the file header in bytes
	// Format: [fileName(32)][entitiesCount(4)][tombstoneCount(4)][nextId(4)] = 44 bytes
	HeaderSize = FileNameSize + (HeaderFieldSize * 3)

	// DefaultBTreeOrder is the default order for B+ tree indices
	DefaultBTreeOrder = 4
)
