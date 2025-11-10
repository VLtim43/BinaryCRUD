(Relatório TPII https://github.com/VLtim43/BinaryCRUD/blob/master/relatorioTPII.md e TPIII https://github.com/VLtim43/BinaryCRUD/blob/master/relatorioTPIII.md )

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

### 2. Install Wails https://wails.io/docs/gettingstarted/installation

Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Running the Application

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

## Development

Built with:

- **Backend**: Go 1.25+
- **Frontend**: Preact + TypeScript
- **Framework**: Wails v2
- **Data Structure**: B+ Tree (order 4)
