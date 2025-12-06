#!/bin/bash
#
# TFDrift-Falco Grafana Quick Start Script
# Automatically sets up and opens Grafana dashboards
#

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "=========================================="
echo "  TFDrift-Falco Grafana Quick Start"
echo "=========================================="
echo ""

# Check if Docker is running
echo -e "${BLUE}[1/4]${NC} Checking Docker..."
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running${NC}"
    echo ""
    echo "Please start Docker Desktop and run this script again."
    exit 1
fi
echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Start Grafana stack
echo -e "${BLUE}[2/4]${NC} Starting Grafana stack..."
cd "$SCRIPT_DIR"
docker-compose up -d

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Grafana stack started${NC}"
else
    echo -e "${RED}✗ Failed to start Grafana stack${NC}"
    exit 1
fi
echo ""

# Wait for services to be ready
echo -e "${BLUE}[3/4]${NC} Waiting for services to be ready..."

# Wait for Grafana
MAX_ATTEMPTS=30
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if curl -s -f http://localhost:3000/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Grafana is ready${NC}"
        break
    fi

    if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
        echo -e "${RED}✗ Grafana did not start in time${NC}"
        echo ""
        echo "Check logs with: docker-compose logs grafana"
        exit 1
    fi

    echo -n "."
    sleep 2
    ATTEMPT=$((ATTEMPT + 1))
done
echo ""

# Open browser
echo -e "${BLUE}[4/4]${NC} Opening Grafana in your browser..."
sleep 2

# Detect OS and open browser
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open http://localhost:3000
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    if command -v xdg-open > /dev/null; then
        xdg-open http://localhost:3000
    elif command -v gnome-open > /dev/null; then
        gnome-open http://localhost:3000
    else
        echo -e "${YELLOW}Please open http://localhost:3000 manually${NC}"
    fi
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    # Windows
    start http://localhost:3000
else
    echo -e "${YELLOW}Please open http://localhost:3000 manually${NC}"
fi

echo ""
echo "=========================================="
echo -e "  ${GREEN}Setup Complete!${NC}"
echo "=========================================="
echo ""
echo "Grafana URL: http://localhost:3000"
echo "Username:    admin"
echo "Password:    admin"
echo ""
echo "Next steps:"
echo "  1. Login to Grafana"
echo "  2. Navigate to Dashboards → TFDrift-Falco"
echo "  3. Explore the sample data"
echo ""
echo "To connect to real TFDrift-Falco data, see:"
echo "  ${SCRIPT_DIR}/GETTING_STARTED.md"
echo ""
echo "Useful commands:"
echo "  View logs:        docker-compose logs -f"
echo "  Stop services:    docker-compose down"
echo "  Restart services: docker-compose restart"
echo ""
