#!/bin/bash
# ControlCRUD - Stop Script
# Stops all services using Docker Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Stopping ControlCRUD...${NC}"

# Stop all services
docker compose down

echo -e "${GREEN}ControlCRUD stopped.${NC}"
echo ""
echo -e "To remove all data (including database), run:"
echo -e "  docker compose down -v"
echo ""
