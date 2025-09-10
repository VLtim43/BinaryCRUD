# BinaryCRUD

A simple binary file CRUD application built with Avalonia UI and .NET 9, demonstrating binary data persistence with a clean architecture pattern.

## Project Structure

### Core Architecture

- **Models**: Data entities and persistence layer
- **ViewModels**: MVVM pattern implementation with business logic
- **Views**: Avalonia UI components and user interface
- **Services**: Cross-cutting concerns and utilities

### Key Components

#### Models Layer (`/Models`)

**Entities**

- `Item.cs` - Core data entity implementing binary serialization

**Interfaces**

- `InterfaceSerializable.cs` - Contract for binary serialization (`ToBytes()`, `FromBytes()`)
- `InterfaceFileDAO.cs` - Data access operations contract

**Base Classes**

- `FileBinaryDAO<T>.cs` - Generic binary file operations with thread safety
- `FileHeader.cs` - File metadata structure (count, timestamps)

**Data Access Objects**

- `ItemDAO.cs` - Item-specific database operations extending base DAO

#### ViewModels Layer (`/ViewModels`)

- `ViewModelBase.cs` - Base class for all view models
- `MainWindowViewModel.cs` - Main application logic and commands

#### Views Layer (`/Views`)

- `MainWindow.axaml` - Main application window with sidebar layout

#### Services Layer (`/Services`)

- `ToastService.cs` - Toast notification system with different message types

### Features

- **Binary File Persistence**: Items stored in binary format with header metadata
- **CRUD Operations**: Create, read, update, delete operations on items
- **Tombstone Pattern**: Soft delete implementation (items marked as deleted but preserved)
- **Real-time UI**: Sidebar automatically refreshes when items are added
- **File Management**: Complete file deletion capability
- **Toast Notifications**: Success (green) and warning (orange) notifications
- **Confirmation Dialogs**: Safety prompts for destructive operations
