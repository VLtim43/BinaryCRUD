package persistence

// ReadAllItems reads all active item records from the binary file.
func ReadAllItems(filename string) ([]Item, error) {
	return readRecords(
		filename,
		"item record",
		ReadItemRecord,
		nil,
		func(item *Item) bool {
			return !item.Tombstone
		},
	)
}
