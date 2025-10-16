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
- **Read**: View all orders
- **Delete**: Remove orders

### Debug Tab

- **Populate Inventory**: Bulk import from JSON file
- **Print Index**: View B+ tree index structure
- **Rebuild Index**: Reconstruct index from data file
- **Delete All Files**: Clear all data
