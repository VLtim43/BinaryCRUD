# BinaryCRUD

A binary file-based CRUD application with B+ tree indexing built with Wails (Go + Preact).

## Running the Built Binary

Download the binary from [Releases](https://github.com/YourUsername/BinaryCRUD/releases) and run:

```bash
# Linux
chmod +x BinaryCRUD
./BinaryCRUD

# Windows
BinaryCRUD.exe
```

## Running it Locally

- install go <https://go.dev/doc/install>
- install wails <https://wails.io/docs/gettingstarted/installation/> (attention to the PATH)
- It also may need some extra libs like libwebkit. I think node and npm too
- Can run it with ./run.sh

## Usage

### Item Tab

- **Create**: Add items to the binary database
- **Read**: Look up items by record ID or print entire file
- **Delete**: Soft-delete items (marks tombstone flag)

### Order Tab

- **Create**: Build orders from existing items (shopping cart)
- **Read**: Look up orders by record ID or print all orders
- **Delete**: Remove orders

### Debug Tab

- **Populate Inventory**: Bulk import from JSON file
- **Print Index**: View B+ tree index structure
- **Rebuild Index**: Reconstruct index from data file
- **Delete All Files**: Clear all data

## Binary Format

### Separators

- **Record Separator (RS)**: `0x1E` - Separates records
- **Unit Separator (US)**: `0x1F` - Separates fields within a record

### Item Record Structure

```
FILE HEADER:
[RecordCount:4 bytes][0x1E]

ITEM RECORD:
[RecordID:4 bytes][0x1F][Tombstone:1 byte][0x1F][NameLength:4 bytes][0x1F][NameData:N bytes][0x1F][Timestamp:8 bytes][0x1E]
```

**Field Details:**

- `RecordCount`: 4 bytes - uint32 little-endian (number of records in file)
- `RecordID`: 4 bytes - uint32 little-endian (unique record ID, assigned from Count before incrementing)
- `Tombstone`: 1 byte - `0x00` = active, `0x01` = deleted
- `NameLength`: 4 bytes - uint32 little-endian (length of name string)
- `NameData`: N bytes - UTF-8 encoded string
- `Timestamp`: 8 bytes - int64 little-endian Unix timestamp

**Example:** Item "pizza" with RecordID=0 (active)

```
Header: [01 00 00 00][0x1E]  (count=1)
Record: [00 00 00 00][0x1F][00][0x1F][05 00 00 00][0x1F][70 69 7A 7A 61][0x1F][<8 bytes timestamp>][0x1E]
```

**ID System:**
- RecordIDs are stored in the record and are incremental (0, 1, 2, 3, ...)
- RecordIDs are assigned from the current Count value before incrementing
- RecordIDs are never reused, even after deletion
- Deleted records remain in the file with tombstone=1
- This allows safe compaction in the future (removing tombstoned records)

### Order Record Structure

```
FILE HEADER:
[RecordCount:4 bytes][0x1E]

ORDER RECORD:
[Tombstone:1 byte][0x1F][ItemCount:4 bytes][0x1F][Items<>][0x1F][Timestamp:8 bytes][0x1E]

EACH ORDER ITEM:
[ItemID:4 bytes][0x1F][Quantity:4 bytes][0x1F]
```

**Field Details:**

- `Tombstone`: 1 byte - `0x00` = active, `0x01` = deleted
- `ItemCount`: 4 bytes - uint32 little-endian (number of items in order)
- `ItemID`: 4 bytes - uint32 little-endian (reference to item record ID)
- `Quantity`: 4 bytes - uint32 little-endian
- `Timestamp`: 8 bytes - int64 little-endian Unix timestamp

**Example:** Order with 2 items (ItemID=1, Qty=3) and (ItemID=5, Qty=1)

```
[00][0x1F][02 00 00 00][0x1F]
  [01 00 00 00][0x1F][03 00 00 00][0x1F]
  [05 00 00 00][0x1F][01 00 00 00][0x1F]
[0x1F][<8 bytes timestamp>][1E]
```

### Index File Structure

The B+ tree index (`.idx` files) maps `RecordID` → `FileOffset` for O(log n) lookups instead of sequential scans.
