# BinaryCRUD

A restaurant manager application built with Wails (Go + Preact) featuring custom binary file storage with B+ tree indexing.

## Features

- **Binary file-based storage** with custom serialization format
- **B+ Tree indexing** (order 4) for fast lookups on:
  - `items.bin` → `items.idx`
  - `orders.bin` → `orders.idx`
  - `promotions.bin` → `promotions.idx`
- **N:N relationship** between Orders and Promotions via `order_promotions.bin`
- **Logical deletion** using tombstone markers (no physical deletion)
- **Wails desktop app** with Go backend and Preact frontend

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

## Data Storage

The application stores data in the `/data` directory:

**Binary files with B+ Tree indexes:**

- `items.bin` / `items.idx` - Menu items
- `orders.bin` / `orders.idx` - Customer orders
- `promotions.bin` / `promotions.idx` - Promotional deals

**Relationship tables:**

- `order_promotions.bin` - N:N relationship

**Binary format:**

- Fixed header: `[entitiesCount(4)][tombstoneCount(4)][nextId(4)]` (12 bytes)
- Length-prefixed records: `[recordLength(2)][recordData...]`
- Tombstone-based logical deletion
