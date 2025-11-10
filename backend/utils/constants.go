package utils

const (
	// UnitSeparator is ASCII 31 (0x1F) - separates fields within a record
	UnitSeparator = "\x1f"

	// RecordSeparator is ASCII 30 (0x1E) - separates entries/records
	RecordSeparator = "\x1e"

	// IDSize is the size of the ID field in bytes
	IDSize = 2

	// TombstoneSize is the size of the tombstone field in bytes
	TombstoneSize = 1

	// HeaderFieldSize is the size of each header field in bytes
	HeaderFieldSize = 4
)
