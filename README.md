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
- **Delete All Files**: Clear all generated data files (`items.bin`, `orders.bin`, `promotions.bin`)

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

## Planned Features

- **Delete Operations**: Implement tombstone-based logical deletion for items, orders, and promotions
- **B+ Tree Indexing**: Fast lookups using B+ tree index structure
- **Update Operations**: Modify existing records (currently only create and read are implemented)
- **Search Functionality**: Search items, orders, and promotions by various criteria
- **Export/Import**: Export data to JSON/CSV formats
- **File Compaction**: Remove tombstoned records to reduce file size
- **Order History**: View and analyze order patterns
- **Promotion Analytics**: Track promotion usage and effectiveness

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
- Item ID (auto-increment per `items.bin`)
- Order ID (auto-increment per `orders.bin`)
- Promotion ID (auto-increment per `promotions.bin`)

**Current Status**: Only primary keys implemented. Each file maintains independent ID sequence in header's `NextID` field.

**Planned**: Secondary indexes for names, dates, and composite keys.

### e) Index Structures (Estruturas de Índice)

**Current Implementation**: Sequential search using `utils.SequentialRead()`

**Planned Structures**:
- **B+ Tree**: For primary key lookups (referenced in Debug tools - Print Index, Rebuild Index)
- **Hash Index**: For exact-match searches on names
- **Extensible Hashing**: For dynamic scaling

See [Planned Features](#planned-features) for B+ tree indexing roadmap.

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

**Current Status**: Not yet implemented.

**Planned Design**:
- **Format**: Separate `.idx` files (e.g., `items.idx`, `orders.idx`)
- **Structure**: B+ tree nodes serialized to disk
- **Updates**: Write-ahead approach - update data file, then rebuild/update index
- **Synchronization**: Index rebuild command available in Debug tab
- **Recovery**: Rebuild Index function to reconstruct from data files

See Debug Tools section for index management commands (planned).

### h) Project Structure (Estrutura do Projeto)

See [Project Structure](#project-structure) section for detailed folder organization.

**Key Architecture Patterns**:

```
├── backend/dao/          # Data Access Objects (one per entity)
│   ├── item_dao.go       # CRUD for items
│   ├── order_dao.go      # CRUD for orders
│   └── promotion_dao.go  # CRUD for promotions
├── backend/utils/        # Shared binary file utilities
│   ├── header.go         # Header read/write/update
│   ├── read.go           # Sequential record reading
│   ├── write.go          # Field encoding (fixed/variable)
│   ├── append.go         # Record appending
│   └── print.go          # Debug printing
├── frontend/src/         # Preact UI components
│   ├── app.tsx           # Main component with tabs
│   └── main.tsx          # Entry point
└── data/                 # Runtime-generated binary files
    ├── items.bin
    ├── orders.bin
    └── promotions.bin
```

**Layered Architecture**:
- **Presentation**: Preact components (frontend/)
- **Application**: Wails bindings (app.go)
- **Business**: DAO layer (backend/dao/)
- **Data Access**: Utils layer (backend/utils/)
- **Storage**: Binary files (data/)

See [Architecture](#architecture) section for data flow diagram.

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]
