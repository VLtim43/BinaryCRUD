# BinaryCRUD

A restaurant manager application built with Wails (Go + Preact) featuring custom binary file storage with B+ tree indexing.

## Prerequisites

### 1. Install Go

Download and install Go from the official website:

- https://go.dev/doc/install

Verify installation:

```bash
go version
```

### 2. Install Wails

Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify installation:

```bash
wails doctor
```

For detailed Wails installation instructions, visit:

- https://wails.io/docs/gettingstarted/installation

## Running the Application

From the project root directory:

```bash
./run.sh
```

Or manually:

```bash
wails dev
```

## Running Tests

Run all backend tests:

```bash
./test.sh
```

Or manually:

```bash
cd backend/test
go test -v
```

## Project Structure

```
BinaryCRUD/
     backend/
        dao/          # Data access layer
            item_dao.go          # Item operations
            collection_dao.go    # Shared order/promotion logic
            order_dao.go         # Order wrapper
            promotion_dao.go     # Promotion wrapper
        index/        # B+ tree implementation
        utils/        # Binary file utilities
    test/         # Test files
     frontend/
    src/          # Preact frontend
     data/             # Binary data files
        items.bin     # Item records
        items.idx     # B+ tree index
        orders.bin    # Order records
        promotions.bin # Promotion records
    items.json    # Sample data
     app.go            # Backend API
     main.go           # Application entry point
     run.sh            # Run the app
 test.sh           # Run tests
```

## Features

- **Binary File Storage**: Custom binary format with tombstone-based logical deletion
- **B+ Tree Indexing**: Fast O(log n) lookups vs O(n) sequential scans for items
- **CRUD Operations**: Create, Read, Delete for items, orders, and promotions
- **Orders**: Create orders with customer names and multiple items
- **Promotions**: Create promotions with multiple items
- **Index Toggle**: Switch between indexed and sequential reads for items
- **Atomic Writes**: File synchronization for data durability
- **Thread-Safe**: Mutex-protected concurrent operations

## How It Works

### Binary Format

**Header (15 bytes):**

```
[entitiesCount(4)][0x1F][tombstoneCount(4)][0x1F][nextId(4)][0x1E]
```

**Item Entry:**

```
[ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][price(4)][0x1E]
```

**Order/Promotion Entry:**

```
[ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)][itemID2(2)]...[0x1E]
```

- Orders use customer name
- Promotions use promotion name
- Both store arrays of item IDs

### Index File

Maps Item ID -> File Offset for fast lookups:

```
[count(8)]
[id(8), offset(8)]...
```

Only items use indexing. Orders and promotions use sequential scan.

## Development

Built with:

- **Backend**: Go 1.25+
- **Frontend**: Preact + TypeScript
- **Framework**: Wails v2
- **Data Structure**: B+ Tree (order 4)
