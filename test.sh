#!/bin/bash

set -e

# Get absolute path to root directory
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ==========================================
# Backend Tests (Go)
# ==========================================
echo -e "${YELLOW}Running Go backend tests...${NC}\n"

if ! go test -v ./backend/test/...; then
    echo -e "\n${RED}✗ Backend tests failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Backend tests passed${NC}\n"

echo -e "${YELLOW}Starting Wails dev server...${NC}"

# Start wails dev in background (cleanup on exit)
WAILS_LOG=$(mktemp)
wails dev -ldflags "-X main.CleanupOnExit=true" > "$WAILS_LOG" 2>&1 &
WAILS_PID=$!

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Stopping Wails dev server...${NC}"
    kill $WAILS_PID 2>/dev/null || true
    wait $WAILS_PID 2>/dev/null || true
    rm -f "$WAILS_LOG"
}
trap cleanup EXIT

# Wait for the dev server to be ready (check if port 34115 is listening)
echo "Waiting for dev server to be ready..."
MAX_WAIT=60
WAITED=0
while ! nc -z localhost 34115 2>/dev/null; do
    sleep 1
    WAITED=$((WAITED + 1))
    if [ $WAITED -ge $MAX_WAIT ]; then
        echo -e "${RED}Timeout waiting for dev server to start${NC}"
        exit 1
    fi
done

echo -e "${GREEN}Dev server is ready!${NC}\n"

# Run frontend tests
echo -e "${YELLOW}Running Playwright frontend tests...${NC}\n"

cd frontend
if ! npx playwright test --reporter=list; then
    echo -e "\n${RED}✗ Frontend tests failed${NC}"
    exit 1
fi

echo -e "\n${GREEN}✓ All tests passed${NC}"
