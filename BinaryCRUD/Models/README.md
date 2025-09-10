# Binary File Format Documentation

This document describes the binary file format used for storing data in the BinaryCRUD application.

## File Structure Overview

Each binary file consists of:

1. **File Header** (12 bytes)
2. **Item Entries** (variable length)

## File Header (12 bytes)

| Offset | Size | Type  | Description                           |
| ------ | ---- | ----- | ------------------------------------- |
| 0      | 4    | int32 | Count - Total number of items in file |
| 4      | 8    | int64 | LastUpdated - Timestamp ticks (UTC)   |

**Total Header Size: 12 bytes**

## Item Entry Structure

Each item entry consists of:

1. **Entry Length Prefix** (4 bytes)
2. **Item Data** (variable length)

### Entry Length Prefix (4 bytes)

| Offset | Size | Type  | Description                                  |
| ------ | ---- | ----- | -------------------------------------------- |
| 0      | 4    | int32 | Total byte length of the following item data |

### Item Data Structure

| Offset | Size | Type    | Description                                           |
| ------ | ---- | ------- | ----------------------------------------------------- |
| 0      | 1    | byte    | **Tombstone bit**: `00` = active, `01` = deleted      |
| 1      | 8    | int64   | **ID**: Unique item identifier                        |
| 9      | 4    | int32   | **Content Length**: Number of bytes in content string |
| 13     | var  | UTF-8   | **Content**: Item name/description                    |
| 13+len | 16   | decimal | **Price**: Item price (4 × int32 values)              |

**Minimum Item Data Size: 29 bytes** (1 + 8 + 4 + 0 + 16)
**Actual Item Data Size: 29 + content_length bytes**

## HEX Example

### Sample File with 1 Active Item

```
File Header (12 bytes):
01000000b4f505abb6f0dd08
│       │
│       └─ LastUpdated: 0x08ddf0b6ab05f5b4 (timestamp ticks)
└─ Count: 1 (0x00000001 little-endian)

Item Entry:
22000000 + 00 + 0100000000000000 + 05000000 + 6170706c65 + 0f000000000000000000000000000100
│        │  │   │                │          │            │
│        │  │   │                │          │            └─ Price: 15.0 (decimal)
│        │  │   │                │          └─ Content: "apple" (UTF-8)
│        │  │   │                └─ Content Length: 5
│        │  │   └─ ID: 1
│        │  └─ Tombstone: 00 (active)
│        └─ Entry Length: 34 bytes (0x22)
└─ Entry Length Prefix
```

### Reading the Tombstone Bit

The **first byte** of each item's data indicates its status:

- `00` = **Active** item
- `01` = **Deleted** item (tombstone)

This allows for soft deletion where items are marked as deleted but not physically removed from the file.

## Implementation Notes

- All multi-byte integers use **little-endian** encoding
- Strings are encoded in **UTF-8**
- Decimal prices are stored as 4 consecutive int32 values (16 bytes total)
- The tombstone mechanism enables soft deletion without file compaction
- Item IDs are assigned sequentially based on the header count
