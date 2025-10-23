# B+ Tree Index Implementation

This package implements a B+ tree index for fast item lookups in BinaryCRUD.

## Overview

The B+ tree index provides O(log n) search complexity for item retrieval by ID, compared to O(n) sequential search. The index maps Item IDs (uint32) to file offsets (int64) in the `items.bin` file.

## Architecture

### Components

1. **bplustree.go** - Core B+ tree implementation
   - `BPlusTree` - Main tree structure with insert/search operations
   - `BPlusNode` - Node structure (internal and leaf nodes)
   - Persistence methods (SaveToFile/LoadFromFile)

2. **item_index.go** - Integration layer
   - `ItemIndex` - Manager for the item index
   - Index lifecycle management (Load/Save/Rebuild)
   - Integration with ItemDAO

## B+ Tree Properties

- **Order**: 4 (configurable)
- **Node capacity**: 3 keys maximum (order - 1)
- **Leaf nodes**: Contain keys and file offsets, linked for range queries
- **Internal nodes**: Contain keys and child pointers for navigation

## File Format

### Index File Structure (`items.bin.idx`)

```
[TreeOrder(4)][EntryCount(4)][Entry1][Entry2]...[EntryN]
```

Each entry:
```
[Key(4)][Offset(8)]
```

- **TreeOrder**: uint32 - B+ tree order
- **EntryCount**: uint32 - Number of entries
- **Key**: uint32 - Item ID
- **Offset**: int64 - File offset in items.bin where record starts

All fields are stored in **little-endian** format.

## Usage

### Integration with ItemDAO

The index is automatically managed by ItemDAO:

**On Write:**
```go
// 1. Get file offset before writing
recordOffset := fileInfo.Size()

// 2. Write record to items.bin
// ...

// 3. Update index
dao.index.Insert(itemID, recordOffset)
dao.index.Save()
```

**On Read with Index:**
```go
// 1. Load index
dao.index.Load()

// 2. Search for offset
offset, found := dao.index.Search(itemID)

// 3. Seek to offset and read record
file.Seek(offset, 0)
// read record...
```

### Rebuilding the Index

The index can be rebuilt from the data file:

```go
dao.RebuildIndex()
```

This scans `items.bin` sequentially and reconstructs the entire B+ tree.

## Operations

### Insert

**Complexity**: O(log n)

Inserts a new key-value pair into the tree. Handles node splitting when nodes reach capacity.

```go
tree.Insert(itemID, fileOffset)
```

### Search

**Complexity**: O(log n)

Finds the file offset for a given item ID.

```go
offset, found := tree.Search(itemID)
```

### Get All Entries

**Complexity**: O(n)

Returns all entries in sorted order by traversing leaf nodes.

```go
entries := tree.GetAllEntries()
```

### Persistence

**Save:**
```go
tree.SaveToFile("items.bin.idx")
```

**Load:**
```go
tree, err := LoadFromFile("items.bin.idx")
```

## B+ Tree Structure Example

For items with IDs [0, 1, 2, 3, 4] (order 4):

```
Internal: keys=[2]
  Leaf: keys=[0, 1], offsets=[14, 30]
  Leaf: keys=[2, 3, 4], offsets=[46, 62, 78]
```

- Root is internal node with pivot key 2
- Left leaf contains items 0-1
- Right leaf contains items 2-4
- Leaf nodes are linked for sequential traversal

## Performance

### Time Complexity

| Operation | Sequential Search | B+ Tree Index |
|-----------|-------------------|---------------|
| Insert    | O(1)*            | O(log n)      |
| Search    | O(n)             | O(log n)      |
| Delete    | O(n)             | O(log n)      |

*Insert at end of file, but index update is O(log n)

### Space Complexity

- **Index file size**: ~12 bytes per entry (4 bytes key + 8 bytes offset)
- **Memory**: O(n) for tree structure

### Benchmarks (Expected)

For 10,000 items:
- Sequential search: ~10,000 comparisons
- B+ tree search: ~log₄(10,000) = ~7 node visits

## Index Lifecycle

### Initialization

1. ItemDAO creates index instance with `.idx` file path
2. Index loaded on first operation (lazy loading)
3. Empty tree created if no index file exists

### Synchronization

- Index updated on every Write operation
- Index persisted to disk after each update
- Rebuild available if index becomes corrupted

### Consistency

- Index and data file can become out of sync if:
  - Write succeeds but index update fails
  - Manual editing of data file
  - Corruption of index file

**Solution**: Use Rebuild Index to reconstruct from data file.

## Frontend Integration

The UI provides:
- **Checkbox**: "Use B+ Tree Index" in Item → Read tab
- **Automatic indexing**: All new items automatically indexed
- **Debug tools**: Print Index and Rebuild Index buttons

## Debugging

### Print Index Structure

```go
fmt.Println(dao.PrintIndex())
```

Output shows tree structure with keys and offsets:
```
Internal: keys=[2]
  Leaf: keys=[0, 1], offsets=[14, 30]
  Leaf: keys=[2, 3, 4], offsets=[46, 62, 78]
```

### Verify Index

Check entry count matches data file:
```go
entries := dao.index.GetAllEntries()
fmt.Printf("Index has %d entries\n", len(entries))
```

## Limitations

- **No deletion support**: Deleted items remain in index (tombstone handling planned)
- **No range queries**: Leaf links exist but range API not exposed
- **Single-threaded**: No concurrent access protection
- **Items only**: Orders and promotions not indexed

## Future Enhancements

- **Deletion**: Remove entries from index on logical deletion
- **Range queries**: Leverage leaf links for efficient range scans
- **Concurrent access**: Add read-write locks for thread safety
- **Secondary indexes**: Index on item names, not just IDs
- **Bulk operations**: Batch inserts for better performance
- **Index compression**: Store offsets as deltas to reduce file size

## References

- [B+ Tree Wikipedia](https://en.wikipedia.org/wiki/B%2B_tree)
- [Database Systems: The Complete Book](http://infolab.stanford.edu/~ullman/dscb.html)

## Testing

The B+ tree implementation can be tested with:

```go
tree := NewBPlusTree(4)
tree.Insert(0, 14)
tree.Insert(1, 30)
offset, found := tree.Search(1)
// offset should be 30, found should be true
```

See ItemDAO integration tests for full coverage.
