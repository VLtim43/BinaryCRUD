# BinaryCRUD

A restaurant management desktop application with custom binary file-based storage, built with Wails (Go + Preact).

BinaryCRUD implements a custom database-like system using binary files with B+ tree indexing and logical deletion with tombstones. It manages items, orders, and promotions using a custom binary file format, providing a modern desktop UI for all CRUD operations.

## Tech Stack

- **Backend**: Go 1.x
- **Frontend**: Preact + TypeScript
- **Framework**: Wails v2
- **Storage**: Custom binary file format with B+ tree indexing

## Getting Started

### Prerequisites

- Go 1.18+
- Node.js 16+
- Wails CLI v2

### Installation

1. Install Go
   https://go.dev/doc/install

2. Install Wails CLI
    https://wails.io/docs/gettingstarted/installation :

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

3. Clone the repository:

```bash
git clone <repository-url>
cd BinaryCRUD
```

4. Install dependencies:

```bash
  go mod download
```

5. Run with

```bash
run.sh
```

## Features

### Core Operations

- **Create**: Add items, orders, and promotions with auto-increment IDs
- **Read**: Retrieve records by ID (with B+ tree indexing for items) or view all records
- **Delete**: Logical deletion using tombstone flags (preserves data)
- **Print**: View complete binary file structure

### Debug Tools

- **Populate Inventory**: Bulk import items from `inventory.json`
- **Print Index**: Display B+ tree structure in console
- **Rebuild Index**: Reconstruct B+ tree index from data file
- **Delete All Files**: Clear all generated data files

## Binary File Format

All three files (`items.bin`, `orders.bin`, `promotions.bin`) share the same header structure with independent ID counters.

### Header Structure (14 bytes)

```
[EntryCount(4)][0x1F][TombstoneCount(4)][0x1F][NextID(4)][0x1E]
```

- **EntryCount**: Number of active records
- **TombstoneCount**: Number of deleted records
- **NextID**: Next auto-increment ID (independent per file)

### Item Record Format

```
[ID(4)][0x1F][Tombstone(1)][0x1F][StringLength(2)][0x1F][StringContent][0x1F][PriceInCents(4)][0x1F][0x1E]
```

- **Tombstone**: `0x00` = active, `0x01` = deleted
- **PriceInCents**: 4-byte integer storing price in cents (e.g., 100 = $1.00)

### Order Record Format

```
[ID(4)][0x1F][Tombstone(1)][0x1F][ItemCount(2)][0x1F][Item1Name][Item2Name]...[0x1E]
```

- **Tombstone**: `0x00` = active, `0x01` = deleted
- Each item name uses variable-length format: `[Length(2)][0x1F][Content][0x1F]`

### Promotion Record Format

```
[ID(4)][0x1F][Tombstone(1)][0x1F][NameLength(2)][0x1F][Name][0x1F][ItemCount(2)][0x1F][Item1Name][Item2Name]...[0x1E]
```

- **Tombstone**: `0x00` = active, `0x01` = deleted
- Each item name uses variable-length format: `[Length(2)][0x1F][Content][0x1F]`

### Separators

- **Unit Separator**: `0x1F` (U+001F) - Separates fields within a record
- **Record Separator**: `0x1E` (U+001E) - Marks the end of a record

## Usage

### Item Management

**Creating Items:**

1. Navigate to **Item** → **Create** tab
2. Enter item name and price in the input fields
3. Click **Add Item**

**Reading Items:**

1. Navigate to **Item** → **Read** tab
2. Enter a record ID
3. Click **Get Record** to fetch a specific item
4. Click **Print** to view the complete binary file in the console

**Deleting Items:**

1. Navigate to **Item** → **Delete** tab
2. Enter the record ID to delete
3. Click **Delete Record** to mark the item as deleted (tombstone flag)

### Order Management

**Creating Orders:**

1. Navigate to **Order** → **Create** tab
2. Select items from the dropdown and click **Add**
3. Adjust quantities as needed (each item shows unit price and total)
4. View the cart total at the top
5. Click **Submit Order** when ready

**Reading Orders:**

1. Navigate to **Order** → **Read** tab
2. Enter an order ID
3. Click **Get Order** to fetch a specific order
4. Click **Print** to view all orders in the console

**Deleting Orders:**

1. Navigate to **Order** → **Delete** tab
2. Enter the order ID to delete
3. Click **Delete Order** to mark the order as deleted (tombstone flag)

### Promotion Management

**Creating Promotions:**

1. Navigate to **Promotion** → **Create** tab
2. Enter a promotion name
3. Select items from the dropdown and click **Add**
4. Adjust quantities as needed (each item shows unit price and total)
5. View the promotion total at the top
6. Click **Create Promotion** when ready

**Reading Promotions:**

1. Navigate to **Promotion** → **Read** tab
2. Enter a promotion ID
3. Click **Get Promotion** to fetch a specific promotion
4. Click **Print** to view all promotions in the console

**Deleting Promotions:**

1. Navigate to **Promotion** → **Delete** tab
2. Enter the promotion ID to delete
3. Click **Delete Promotion** to mark the promotion as deleted (tombstone flag)

### Populating Test Data

1. Create an `inventory.json` file in the project root:

```json
{
  "items": [
    { "name": "Pizza", "priceInCents": 1200 },
    { "name": "Burger", "priceInCents": 899 },
    { "name": "Chicken Wrap", "priceInCents": 750 },
    { "name": "Burrito", "priceInCents": 950 },
    { "name": "Salad", "priceInCents": 650 }
  ]
}
```

2. Navigate to **Debug** tab
3. Click **Populate Inventory**

## Architecture

**Layered Structure:**

```
UI (Preact) → Wails Bindings → App Layer → DAO Layer → Index Layer → Utils Layer → Binary Files
```

- **Presentation**: Tab-based Preact UI with real-time feedback
- **Application**: Wails bindings (app.go)
- **Business**: DAO layer (item_dao.go, order_dao.go, promotion_dao.go)
- **Indexing**: B+ tree layer (bplustree.go, item_index.go)
- **Data Access**: Utils layer (header, read, write, append, print)
- **Storage**: Binary files with separate index files

## Implementation Status

### ✅ Implemented

- B+ Tree indexing for items (O(log n) lookups)
- Create, Read, Delete operations for all entities
- Logical deletion with tombstone flags
- Index persistence and rebuilding
- Debug tools and UI

### 🚧 Planned

- Update operations
- B+ Tree for orders/promotions
- Secondary indexes
- File compaction
- Concurrent access control

## Technical Implementation Details

### a) Record Structure (Estrutura de Registros)

All records use a combination of **fixed-length** and **variable-length** fields with separators:

- **Fixed-length fields**: IDs (4 bytes), counts (2 bytes) stored in little-endian format
- **Variable-length fields**: Strings encoded as `[Length(2)][0x1F][Content][0x1F]`
- **Separators**: Unit Separator (0x1F) between fields, Record Separator (0x1E) at record end

See [Binary File Format](#binary-file-format) section for detailed format specifications.

### b) Multi-valued String Attributes (Atributos Multivalorados)

Multi-valued string attributes (items in orders/promotions) are stored using:

1. **Count field**: 2-byte field indicating number of strings
2. **Variable-length encoding**: Each string uses `[Length(2)][0x1F][Content][0x1F]`
3. **Sequential storage**: Strings stored one after another

**Example** (Order with 3 items):

```
[ItemCount: 03 00][0x1F][Item1][Item2][Item3]
```

See Order and Promotion record formats in [Binary File Format](#binary-file-format).

### c) Logical Deletion Implementation (Exclusão Lógica)

Logical deletion using tombstone flags preserves data while marking records as deleted:

- **Tombstone flag**: 1-byte field per record (0x00 = active, 0x01 = deleted)
- **Header tracking**: `TombstoneCount` tracks total deleted records
- **Delete process**:
  - Items: Use B+ tree index for O(log n) lookup
  - Orders/Promotions: Sequential search
  - Flip tombstone byte from 0x00 to 0x01
  - Increment `TombstoneCount` in header
- **Read operations**: Automatically skip tombstoned records
- **Benefits**: Data preserved for recovery, no file reorganization, fast deletion

**API**: `DeleteItem(id)`, `DeleteOrder(id)`, `DeletePromotion(id)`

### d) Search Keys (Chaves de Pesquisa)

**Primary Keys**: Auto-increment IDs managed in each file's `NextID` header field

- **Item ID** - ✅ **Indexed with B+ Tree** (O(log n) search)
- **Order ID** - Sequential search (O(n))
- **Promotion ID** - Sequential search (O(n))

**Item Index Details**:

- B+ tree (order 4) stored in `items.bin.idx`
- Maps Item ID (uint32) → File Offset (int64)
- Automatic updates on Write operations
- UI toggle: "Use B+ Tree Index" checkbox

**Planned**: Secondary indexes for names, dates, composite keys

### e) Index Structures (Estruturas de Índice)

**B+ Tree Index for Items** (`backend/index/`)

**Specifications**:

- **Type**: B+ Tree (order 4)
- **Key → Value**: Item ID (uint32) → File Offset (int64)
- **File**: `items.bin.idx`
- **Format**: `[TreeOrder(4)][EntryCount(4)][Key(4)+Offset(8)]...`

**Node Structure**:

- **Internal nodes**: Keys + child pointers (max 3 keys)
- **Leaf nodes**: Keys + file offsets, linked for sequential access

**Operations**:

- **Insert**: O(log n) with automatic node splitting
- **Search**: O(log n) with binary search
- **Persistence**: Auto-save after each insert
- **Rebuild**: Reconstruct from data file (`dao.RebuildIndex()`)

**Usage**:

```go
dao.Write("Pizza")              // Creates record + updates index
offset, found := index.Search(id)  // O(log n) lookup
dao.RebuildIndex()              // Recovery from corruption
```

**UI**: Checkbox in Item → Read tab (enabled by default)

**Planned**: Hash indexes, extensible hashing, B+ tree for orders/promotions

See `backend/index/README.md` for complete documentation.

### f) 1:N Relationship Implementation

**Approach**: Embedded collections (denormalized)

Orders and Promotions store item **names** (not IDs) directly in records.

**Trade-offs**:

- ✓ Fast read (no joins)
- ✗ Data duplication
- ✗ No referential integrity
- ✗ No cascade updates

**Example**: Order record contains `["Pizza", "Burrito"]` as embedded strings

**Future**: Store item IDs with join logic for referential integrity

### g) Index Persistence (Persistência de Índices)

**File Format**: `items.bin.idx` (little-endian)

```
[TreeOrder(4)][EntryCount(4)][Entry1][Entry2]...[EntryN]
Entry: [Key(4)][Offset(8)]  // 12 bytes
```

**Persistence Strategy**:

1. **On Write**: Write record → Load index → Insert(ID, offset) → Save index
2. **On Read**: Load index → Search(ID) → Seek to offset → Read record
3. **Synchronization**: Index saved after every Write (atomic operation)

**Index Lifecycle**:

- **Creation**: Auto-created on first Write
- **Updates**: Real-time on every insert
- **Loading**: Lazy load from disk
- **Rebuilding**: Manual via `dao.RebuildIndex()`

**Recovery**: If corrupted, rebuild scans entire data file and reconstructs B+ tree

**Consistency**:

- ✅ Index reflects latest data
- ✅ Rebuild available
- ⚠️ No transactions (write may succeed, index update may fail)

**Performance**: ~12 bytes per item, O(n) load/save time

**UI Commands**: Debug → Print Index / Rebuild Index

### h) Project Structure (Estrutura do Projeto)

```
BinaryCRUD/
├── backend/
│   ├── dao/              # Data Access Objects
│   │   ├── item_dao.go       # CRUD + B+ tree index
│   │   ├── order_dao.go      # CRUD for orders
│   │   └── promotion_dao.go  # CRUD for promotions
│   ├── index/            # B+ tree implementation
│   │   ├── bplustree.go      # Core tree (insert, search, persist)
│   │   ├── item_index.go     # Index manager
│   │   └── README.md         # Documentation
│   └── utils/            # Binary file utilities
│       ├── header.go         # Header operations
│       ├── read.go           # Sequential reading
│       ├── write.go          # Field encoding
│       ├── append.go         # Record appending
│       └── print.go          # Debug printing
├── frontend/src/
│   ├── app.tsx           # Main UI
│   └── main.tsx          # Entry point
├── data/                 # Generated at runtime
│   ├── items.bin
│   ├── items.bin.idx
│   ├── orders.bin
│   └── promotions.bin
├── app.go                # Wails bindings
└── main.go               # Entry point
```

See [Architecture](#architecture) section above for layered structure and data flow.
