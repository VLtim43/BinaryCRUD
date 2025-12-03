# E2E Test Plan

## Test Execution Order

Tests run sequentially (not parallel) because they share state.

### Phase 1: Setup & Basic Validation
- [x] `01-setup.spec.ts` - App loads, tabs visible, delete all, populate inventory

### Phase 2: CRUD Operations (Happy Path)
- [ ] `02-items.spec.ts` - Create, Read, Delete items
- [ ] `03-orders.spec.ts` - Create orders with items, view, delete
- [ ] `04-promotions.spec.ts` - Create promotions, apply to orders

### Phase 3: Validation & Error Handling
- [ ] `05-validation.spec.ts` - Empty names, invalid IDs, missing items

### Phase 4: Advanced Features
- [ ] `06-compression.spec.ts` - Compress files (Huffman, LZW), decompress
- [ ] `07-encryption.spec.ts` - Toggle encryption, verify data integrity
- [ ] `08-compact.spec.ts` - Compact database after deletions

### Phase 5: Cleanup
- [ ] `09-cleanup.spec.ts` - Delete all files at the end

---

## How to Run Tests

### Option 1: Full test suite (backend + frontend)
```bash
./test.sh
```

### Option 2: Manual (for development)

**Terminal 1** - Start Wails:
```bash
cd frontend && npm run wails:dev
# or from root: wails dev
```

**Terminal 2** - Run tests:
```bash
cd frontend

# Headless
npm test

# Watch in browser
npm run test:headed

# Interactive UI
npm run test:ui

# Specific file
npx playwright test tests/e2e/01-setup.spec.ts
```

### View Report
```bash
npm run test:report
```

---

## Test Data

- Tests use the real `data/` folder
- First test deletes all files to start clean
- Populate inventory loads from `data/seed/`
- Last test should clean up (delete all)

---

## Adding New Tests

1. Create file in `tests/e2e/` with numeric prefix for ordering
2. Use `test.describe` to group related tests
3. Tests run in file order (01, 02, 03...)
