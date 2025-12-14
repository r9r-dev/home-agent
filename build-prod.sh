#!/bin/bash

# Home Agent Production Build Script
# This script builds both frontend and backend for production

set -e

echo "🏠 Home Agent - Production Build"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f ".env" ] && [ ! -f "backend/.env" ]; then
    echo -e "${YELLOW}Warning: No .env file found${NC}"
    echo "Make sure to create one before running in production"
    echo ""
fi

# Build frontend
echo -e "${GREEN}[1/3] Building frontend...${NC}"
cd frontend

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm install
fi

# Build
npm run build
echo -e "${GREEN}✓ Frontend built successfully${NC}"
echo ""

# Build backend
echo -e "${GREEN}[2/3] Building backend...${NC}"
cd ../backend

# Check Go dependencies
if [ ! -f "go.sum" ]; then
    echo "Installing Go dependencies..."
    go mod download
fi

# Build with optimizations
go build -ldflags="-s -w" -o home-agent
echo -e "${GREEN}✓ Backend built successfully${NC}"
echo ""

# Summary
echo -e "${GREEN}[3/3] Build complete!${NC}"
echo ""
echo "Output files:"
echo "  - Backend binary: backend/home-agent"
echo "  - Frontend assets: backend/public/"
echo ""
echo "To run in production:"
echo "  cd backend"
echo "  ./home-agent"
echo ""
echo "The server will serve both API and static files on port 8080"
