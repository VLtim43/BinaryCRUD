# BinaryCRUD

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Preact](https://img.shields.io/badge/Preact-673AB8?style=for-the-badge&logo=preact&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-DF0000?style=for-the-badge&logo=wails&logoColor=white)
![Playwright](https://img.shields.io/badge/Playwright-2EAD33?style=for-the-badge&logo=playwright&logoColor=white)

A restaurant manager application built with Wails (Go + Preact) featuring custom binary file storage with B+ tree indexing.

## Features

- **Binary file-based storage** with custom serialization format
- **B+ Tree indexing** (order 4) for fast lookups on:
  - `items.bin` → `items.idx`
  - `orders.bin` → `orders.idx`
  - `promotions.bin` → `promotions.idx`
- **N:N relationship** between Orders and Promotions via `order_promotions.bin`
- **Logical deletion** using tombstone markers (no physical deletion)
- **Compression algorithms:**
  - Huffman coding
  - LZW compression
- **RSA encryption** (2048-bit RSA-OAEP with SHA-256) for sensitive fields
- **Pattern matching search:**
  - KMP (Knuth-Morris-Pratt)
  - Boyer-Moore (Bad Character heuristic)
- **Wails desktop app** with Go backend and Preact frontend

### Prerequisites

- Go 1.18+
- Node.js 16+
- Wails CLI v2

### Installation

1. Install Go (On Windows you may need to restart your computer) and Node/NPM

https://go.dev/doc/install

https://nodejs.org/en/download

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

## Project Structure

```
BinaryCRUD/
├── app.go                 <- Main API: exposes all backend functions to frontend
├── main.go                <- Application entry point
├── logger.go              <- Logging system
├── toast.go               <- Toast notification system
│
├── backend/
│   ├── dao/               <- Data Access Objects (CRUD operations)
│   │   ├── item_dao.go        <- Items table operations + search
│   │   ├── order_dao.go       <- Orders table operations
│   │   ├── promotion_dao.go   <- Promotions table operations
│   │   ├── order_promotion_dao.go  <- N:N relationship operations
│   │   └── collection_dao.go  <- Shared logic for orders/promotions
│   │
│   ├── index/             <- Indexing structures
│   │   ├── btree.go           <- B+ Tree implementation (order 4)
│   │   ├── extensible_hash.go <- Extensible hashing (alternative index)
│   │   └── persistence.go     <- Index serialization to disk
│   │
│   ├── search/            <- Pattern matching algorithms
│   │   ├── kmp.go             <- KMP (Knuth-Morris-Pratt) algorithm
│   │   └── boyer_moore.go     <- Boyer-Moore algorithm
│   │
│   ├── compression/       <- Compression algorithms
│   │   ├── huffman.go         <- Huffman coding
│   │   ├── lzw.go             <- LZW compression
│   │   └── compressor.go      <- Compression interface
│   │
│   ├── crypto/            <- Encryption
│   │   ├── rsa.go             <- RSA-OAEP encryption (production)
│   │   └── simple_rsa.go      <- Educational RSA implementation
│   │
│   ├── utils/             <- Helper functions (read, write, parse, validate, etc.)
│   └── test/              <- Unit tests for all backend components
│
├── frontend/
│   └── src/
│       ├── app.tsx            <- Main React component
│       ├── components/
│       │   ├── tabs/          <- Tab components (ItemTab, OrderTab, etc.)
│       │   └── ...            <- Reusable UI components (Button, Input, etc.)
│       ├── services/          <- API service calls to backend
│       └── utils/             <- Frontend helpers (formatters, toast)
│
└── data/
    ├── bin/               <- Binary data files (.bin)
    ├── indexes/           <- B+ Tree index files (.idx)
    ├── compressed/        <- Compressed files (.huffman, .lzw)
    ├── keys/              <- RSA keys (private.pem, public.pem)
    └── seed/              <- Initial data (items.json, orders.json, etc.)
```

## Database Schema

### Tables

**Items** (`items.bin`)
| Field | Size | Description |
|-------|------|-------------|
| id | 2 bytes | Auto-increment primary key |
| tombstone | 1 byte | 0x00 = active, 0x01 = deleted |
| nameLength | 2 bytes | Length of name string |
| name | variable | Item name (e.g., "Classic Burger") |
| price | 4 bytes | Price in cents (e.g., 899 = $8.99) |

**Orders** (`orders.bin`)
| Field | Size | Description |
|-------|------|-------------|
| id | 2 bytes | Auto-increment primary key |
| tombstone | 1 byte | Deletion marker |
| ownerLength | 2 bytes | Length of customer name |
| owner | variable | Customer name (RSA encrypted) |
| totalPrice | 4 bytes | Total in cents |
| itemCount | 2 bytes | Number of items |
| itemIDs | 2 bytes each | Array of item IDs |

**Promotions** (`promotions.bin`)
| Field | Size | Description |
|-------|------|-------------|
| id | 2 bytes | Auto-increment primary key |
| tombstone | 1 byte | Deletion marker |
| nameLength | 2 bytes | Length of promotion name |
| name | variable | Promotion name (RSA encrypted) |
| totalPrice | 4 bytes | Bundle price in cents |
| itemCount | 2 bytes | Number of items |
| itemIDs | 2 bytes each | Array of item IDs |

**OrderPromotions** (`order_promotions.bin`)
| Field | Size | Description |
|-------|------|-------------|
| orderID | 2 bytes | Foreign key to Orders |
| promotionID | 2 bytes | Foreign key to Promotions |
| tombstone | 1 byte | Deletion marker |

### Relationships

| Relationship        | Type | How                                   |
| ------------------- | ---- | ------------------------------------- |
| Orders ↔ Items      | M:N  | `itemIDs[]` embedded in Orders        |
| Promotions ↔ Items  | M:N  | `itemIDs[]` embedded in Promotions    |
| Orders ↔ Promotions | N:N  | `order_promotions.bin` junction table |

**Many-to-Many Details:**
- **Items ↔ Orders/Promotions**: A single item can appear in multiple orders and multiple promotions. Likewise, each order or promotion can contain multiple items.
- **Orders ↔ Promotions**: An order can include multiple promotions, and a promotion can be applied to multiple orders.
