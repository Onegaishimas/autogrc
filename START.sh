#!/bin/bash
# ControlCRUD - Start Script
# Starts all services using Docker Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting ControlCRUD...${NC}"

# Check if .env file exists, if not create from example
if [ ! -f .env ]; then
    echo -e "${YELLOW}No .env file found. Creating from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        # Generate encryption key if not set
        if grep -q "GENERATE_A_SECURE_KEY_HERE" .env; then
            ENCRYPTION_KEY=$(openssl rand -base64 32)
            sed -i "s|GENERATE_A_SECURE_KEY_HERE|$ENCRYPTION_KEY|g" .env
            echo -e "${GREEN}Generated new encryption key${NC}"
        fi
        echo -e "${YELLOW}Please review .env file and update passwords for production use${NC}"
    else
        echo -e "${RED}Error: .env.example not found${NC}"
        exit 1
    fi
fi

# Create nginx logs directory if it doesn't exist
mkdir -p nginx/logs

# Build and start all services
echo -e "${GREEN}Building and starting containers...${NC}"
docker compose up -d --build

# Wait for services to be healthy
echo -e "${YELLOW}Waiting for services to become healthy...${NC}"
sleep 5

# Check health
echo ""
echo -e "${GREEN}Checking service health...${NC}"
docker compose ps

# Show access information
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}ControlCRUD is running!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Access the application at:"
echo -e "  ${GREEN}http://autogrc.mcslab.io${NC}"
echo -e "  ${GREEN}http://localhost${NC}"
echo ""
echo -e "Health check endpoint:"
echo -e "  ${GREEN}http://localhost/health${NC}"
echo ""
echo -e "View logs:"
echo -e "  docker compose logs -f"
echo ""
echo -e "Stop services:"
echo -e "  ./STOP.sh"
echo ""
