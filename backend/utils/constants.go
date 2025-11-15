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

	// HeaderSize is the total size of the file header in bytes
	// Format: [entitiesCount(4)][tombstoneCount(4)][nextId(4)] = 12 bytes
	HeaderSize = HeaderFieldSize * 3
)
