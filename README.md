# BinaryCRUD

A restaurant management desktop application with custom binary file-based storage, built with Wails (Go + Preact).

## Overview

BinaryCRUD is a restaurant manager application that implements a custom database-like system using binary files with B+ tree indexing and logical deletion with tombstones. It manages items, orders, and promotions using a custom binary file format, providing a modern desktop UI for all CRUD operations.

## Tech Stack

- **Backend**: Go 1.x
- **Frontend**: Preact + TypeScript
- **Framework**: Wails v2
- **Storage**: Custom binary file format

## Features

### Item Management

- **Create**: Add new items to the inventory
- **Read**: Retrieve items by record ID or view all items
- **Delete**: Mark items as deleted using tombstone flags (planned)
- **Print**: View the complete binary file structure

### Order Management

- **Create**: Build orders by selecting multiple items with quantities
- **Read**: Retrieve orders by ID or view all orders
- **Delete**: Mark orders as deleted using tombstone flags (planned)
- **Print**: View the complete orders binary file

### Promotion Management

- **Create**: Create named promotions (collections of items)
- **Read**: Retrieve promotions by ID or view all promotions
- **Delete**: Mark promotions as deleted using tombstone flags (planned)
- **Print**: View the complete promotions binary file

### Debug Tools

- **Populate Inventory**: Bulk import items from `inventory.json`
- **Print Index**: Display B+ tree structure in console (shows keys, offsets, tree levels)
- **Rebuild Index**: Reconstruct B+ tree index from `items.bin` data file
- **Delete All Files**: Clear all generated data files (`items.bin`, `items.bin.idx`, `orders.bin`, `promotions.bin`)

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
[ID(4)][0x1F][StringLength(2)][0x1F][StringContent][0x1F][0x1E]
```

### Order Record Format

```
[ID(4)][0x1F][ItemCount(2)][0x1F][Item1Name][Item2Name]...[0x1E]
```

Each item name uses variable-length format: `[Length(2)][0x1F][Content][0x1F]`

### Promotion Record Format

```
[ID(4)][0x1F][NameLength(2)][0x1F][Name][0x1F][ItemCount(2)][0x1F][Item1Name][Item2Name]...[0x1E]
```

Each item name uses variable-length format: `[Length(2)][0x1F][Content][0x1F]`

### Separators

- **Unit Separator**: `0x1F` (U+001F) - Separates fields within a record
- **Record Separator**: `0x1E` (U+001E) - Marks the end of a record

## Project Structure

```
BinaryCRUD/
├── backend/
│   ├── dao/
│   │   ├── item_dao.go        # Item data access layer
│   │   ├── order_dao.go       # Order data access layer
│   │   └── promotion_dao.go   # Promotion data access layer
│   └── utils/
│       ├── header.go          # Binary file header operations
│       ├── write.go           # Binary write utilities
│       ├── append.go          # Record appending
│       ├── read.go            # Sequential read utilities
│       └── print.go           # Debug printing
├── frontend/
│   └── src/
│       ├── app.tsx            # Main UI component
│       └── main.tsx           # App entry point
├── data/                      # Generated at runtime
│   ├── items.bin              # Items binary file
│   ├── orders.bin             # Orders binary file
│   └── promotions.bin         # Promotions binary file
├── app.go                     # Wails application bindings
├── main.go                    # Application entry point
└── inventory.json             # Sample inventory data
```

## Getting Started

### Prerequisites

- Go 1.18+
- Node.js 16+
- Wails CLI v2

### Installation

1. Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

2. Clone the repository:

```bash
git clone <repository-url>
cd BinaryCRUD
```

3. Install dependencies:

```bash
wails build
```

### Development

Run in development mode with hot reload:

```bash
.\run.sh
```

### Building

Build for production:

```bash
wails build
```

The compiled application will be in the `build/bin` directory.

## Usage

### Item Management

**Creating Items:**

1. Navigate to **Item** → **Create** tab
2. Enter item name in the input field
3. Click **Add Item**

**Reading Items:**

1. Navigate to **Item** → **Read** tab
2. Enter a record ID
3. Click **Get Record** to fetch a specific item
4. Click **Print** to view the complete binary file in the console

**Deleting Items:**

1. Navigate to **Item** → **Delete** tab
2. Enter the record ID to delete
3. Click **Delete Record** (not yet implemented)

### Order Management

**Creating Orders:**

1. Navigate to **Order** → **Create** tab
2. Select items from the dropdown and click **Add**
3. Adjust quantities as needed
4. Click **Submit Order** when ready

**Reading Orders:**

1. Navigate to **Order** → **Read** tab
2. Enter an order ID
3. Click **Get Order** to fetch a specific order
4. Click **Print** to view all orders in the console

**Deleting Orders:**

1. Navigate to **Order** → **Delete** tab
2. Enter the order ID to delete
3. Click **Delete Order** (not yet implemented)

### Promotion Management

**Creating Promotions:**

1. Navigate to **Promotion** → **Create** tab
2. Enter a promotion name
3. Select items from the dropdown and click **Add**
4. Adjust quantities as needed
5. Click **Create Promotion** when ready

**Reading Promotions:**

1. Navigate to **Promotion** → **Read** tab
2. Enter a promotion ID
3. Click **Get Promotion** to fetch a specific promotion
4. Click **Print** to view all promotions in the console

**Deleting Promotions:**

1. Navigate to **Promotion** → **Delete** tab
2. Enter the promotion ID to delete
3. Click **Delete Promotion** (not yet implemented)

### Populating Test Data

1. Create an `inventory.json` file in the project root:

```json
{
  "items": ["Pizza", "Burger", "Chicken Wrap", "Burrito", "Salad"]
}
```

2. Navigate to **Debug** tab
3. Click **Populate Inventory**

## Architecture

### Backend (Go)

**DAO Layer:**

- **ItemDAO**: Handles item persistence (Create, Read, Print)
- **OrderDAO**: Handles order persistence (Create, Read by ID, Print)
- **PromotionDAO**: Handles promotion persistence (Create, Read by ID, Print)

**Utils Package:**

- Header management (EntryCount, TombstoneCount, NextID)
- Variable-length string encoding/decoding
- Fixed-length field encoding (IDs, counts)
- Sequential record reading
- Record appending
- Binary file printing

### Frontend (Preact)

- Tab-based navigation (Item, Order, Promotion, Debug)
- Sub-tabs for each entity (Create, Read, Delete)
- Direct bindings to Go backend functions via Wails
- Real-time feedback for all operations
- Cart-style UI for building orders and promotions

### Data Flow

```
UI (Preact) → Wails Bindings → App Layer (app.go) → DAO Layer → Utils Layer → Binary Files
```

## Features Status

### ✅ Implemented

- **B+ Tree Indexing for Items**: O(log n) fast lookups with automatic index maintenance
- **Create Operations**: Add items, orders, and promotions with auto-increment IDs
- **Read Operations**: Retrieve records by ID (sequential or indexed for items)
- **Index Persistence**: Automatic save/load of B+ tree to `.idx` files
- **Index Rebuilding**: Reconstruct index from data file
- **Debug Tools**: Print binary files, print index structure, rebuild index

### 🚧 Planned Features

- **Delete Operations**: Implement tombstone-based logical deletion for items, orders, and promotions
- **Update Operations**: Modify existing records (currently only create and read are implemented)
- **B+ Tree for Orders/Promotions**: Extend indexing to other entities
- **Secondary Indexes**: Index on names, dates, and composite keys
- **Search Functionality**: Search items, orders, and promotions by various criteria
- **Export/Import**: Export data to JSON/CSV formats
- **File Compaction**: Remove tombstoned records to reduce file size
- **Order History**: View and analyze order patterns
- **Promotion Analytics**: Track promotion usage and effectiveness
- **Concurrent Access**: Add read-write locks for thread safety

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

**Current Status**: Structure in place, implementation planned.

- **Header tracking**: `TombstoneCount` field in file header (see [Header Structure](#header-structure-14-bytes))
- **Planned approach**: Tombstone flag in record to mark as deleted
- **Benefits**: Preserves data for recovery, maintains referential integrity
- **UI**: Delete tabs exist in all sections (Item, Order, Promotion) showing "not yet implemented"

### d) Search Keys (Chaves de Pesquisa)

**Primary Keys (PKs):**

- **Item ID** (auto-increment per `items.bin`) - ✅ **Indexed with B+ Tree**
- **Order ID** (auto-increment per `orders.bin`) - Sequential search only
- **Promotion ID** (auto-increment per `promotions.bin`) - Sequential search only

**Implementation Details:**

Each file maintains independent ID sequence in header's `NextID` field.

**Item ID** is indexed using a B+ tree (order 4) stored in `items.bin.idx`:
- Maps Item ID (uint32) → File Offset (int64)
- O(log n) search complexity
- Automatic index updates on Write operations
- User-selectable in UI: "Use B+ Tree Index" checkbox

**Planned**: Secondary indexes for names, dates, and composite keys for orders/promotions.

### e) Index Structures (Estruturas de Índice)

**✅ Implemented: B+ Tree Index for Items**

**Location**: `backend/index/`
- `bplustree.go` - Core B+ tree implementation
- `item_index.go` - Integration with ItemDAO

**Specifications**:
- **Type**: B+ Tree with order 4
- **Key**: Item ID (uint32)
- **Value**: File offset in items.bin (int64)
- **File**: `items.bin.idx` (separate index file)
- **Format**: `[TreeOrder(4)][EntryCount(4)][Key(4)+Offset(8)]...`

**Features**:
- **Insert**: O(log n) with automatic node splitting
- **Search**: O(log n) with binary search in nodes
- **Persistence**: Automatic save after each insert
- **Rebuild**: Can reconstruct entire index from data file
- **Debugging**: Print tree structure command

**Node Structure**:
- **Internal nodes**: Keys + child pointers for navigation
- **Leaf nodes**: Keys + file offsets, linked for sequential traversal
- **Max keys per node**: 3 (order - 1)

**Usage Example**:

```go
// Automatic on Write
dao.Write("Pizza")  // Creates record + updates index

// Fast lookup
offset, found := index.Search(itemID)  // O(log n)
// vs sequential: O(n)

// Rebuild if needed
dao.RebuildIndex()  // Scans entire data file
```

**UI Integration**:
- Checkbox in Item → Read tab: "Use B+ Tree Index"
- Enabled by default for optimal performance
- Shows search method used in result message

**Orders and Promotions**: Currently use `utils.SequentialRead()` (O(n))

**Planned Structures**:
- **Hash Index**: For exact-match searches on item names
- **Extensible Hashing**: For dynamic scaling
- **B+ Tree for Orders/Promotions**: Extend indexing to other entities

See `backend/index/README.md` for complete B+ tree documentation.

### f) 1:N Relationship Implementation

**Current Approach**: Embedded collection (denormalized)

Orders and Promotions store item **names** (not IDs) as embedded collections:
- **Navigation**: Direct - items are embedded in order/promotion records
- **Referential Integrity**: None currently enforced (items stored by name, not FK)
- **Trade-offs**:
  - ✓ Fast read (no joins needed)
  - ✗ Data duplication (item names repeated)
  - ✗ No cascade updates if item names change

**Example**: Order contains `["Pizza", "Burrito"]` - full names stored, not references to items.bin.

**Future Enhancement**: Could store item IDs instead and implement join logic for referential integrity.

### g) Index Persistence (Persistência de Índices)

**✅ Implemented for Items**

**File Format**: `items.bin.idx`

**Structure**:
```
[TreeOrder(4)][EntryCount(4)][Entry1][Entry2]...[EntryN]

Each entry:
[Key(4)][Offset(8)]  // 12 bytes per entry
```

All fields stored in **little-endian** format.

**Persistence Strategy**:

1. **On Write Operation**:
   - Get current file size (record offset)
   - Write record to data file
   - Load index from disk
   - Insert new entry: `index.Insert(itemID, offset)`
   - Save index to disk

2. **On Read with Index**:
   - Load index from disk (cached in memory)
   - Search B+ tree for offset
   - Seek directly to offset in data file
   - Read single record

3. **Index Synchronization**:
   - Index saved after every Write operation
   - Atomic operation: data write + index update
   - If index update fails, logged as warning (data still written)

**Index Lifecycle**:

- **Creation**: Automatically created on first Write
- **Updates**: Real-time updates on every item insert
- **Loading**: Lazy loading - read from disk on first operation
- **Rebuilding**: Manual rebuild command available

**Recovery Mechanism**:

If index becomes corrupted or out of sync:

```go
dao.RebuildIndex()
```

Process:
1. Scans entire `items.bin` file
2. Parses each record to extract ID and offset
3. Builds new B+ tree from scratch
4. Saves reconstructed index

**Consistency Guarantees**:

- ✅ Index always reflects latest data (updated on Write)
- ✅ Rebuild available if corruption occurs
- ✅ Index operations logged for debugging
- ⚠️ No transaction support - write may succeed but index update fail

**Performance**:

- **Index file size**: ~12 bytes per item
- **Load time**: O(n) to rebuild tree from flat file
- **Save time**: O(n) to serialize all entries
- **Memory**: O(n) for in-memory tree structure

**UI Commands**:

- **Print Index**: Debug → Print Index (shows tree structure in console)
- **Rebuild Index**: Debug → Rebuild Index (reconstructs from data file)

See `backend/index/README.md` for implementation details.

### h) Project Structure (Estrutura do Projeto)

See [Project Structure](#project-structure) section for detailed folder organization.

**Key Architecture Patterns**:

```
├── backend/
│   ├── dao/              # Data Access Objects (one per entity)
│   │   ├── item_dao.go       # CRUD + B+ tree index integration
│   │   ├── order_dao.go      # CRUD for orders
│   │   └── promotion_dao.go  # CRUD for promotions
│   ├── index/            # B+ tree index implementation
│   │   ├── bplustree.go      # Core B+ tree (insert, search, persist)
│   │   ├── item_index.go     # Index manager for items
│   │   └── README.md         # Complete index documentation
│   └── utils/            # Shared binary file utilities
│       ├── header.go         # Header read/write/update
│       ├── read.go           # Sequential record reading
│       ├── write.go          # Field encoding (fixed/variable)
│       ├── append.go         # Record appending
│       └── print.go          # Debug printing
├── frontend/src/         # Preact UI components
│   ├── app.tsx           # Main UI with index toggle
│   └── main.tsx          # Entry point
└── data/                 # Runtime-generated files
    ├── items.bin         # Item data file
    ├── items.bin.idx     # B+ tree index for items ✨
    ├── orders.bin        # Order data file
    └── promotions.bin    # Promotion data file
```

**Layered Architecture**:
- **Presentation**: Preact components (frontend/)
- **Application**: Wails bindings (app.go)
- **Business**: DAO layer (backend/dao/)
- **Indexing**: B+ tree layer (backend/index/) - **New!**
- **Data Access**: Utils layer (backend/utils/)
- **Storage**: Binary files (data/)

See [Architecture](#architecture) section for data flow diagram.

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]
