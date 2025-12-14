.PHONY: help build run stop clean dev test

# Colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

# Default target
help:
	@echo '${GREEN}Home Agent - Makefile Commands${RESET}'
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo '  ${YELLOW}build${RESET}       Build Docker image'
	@echo '  ${YELLOW}run${RESET}         Run containers with docker compose'
	@echo '  ${YELLOW}stop${RESET}        Stop running containers'
	@echo '  ${YELLOW}restart${RESET}     Restart containers'
	@echo '  ${YELLOW}logs${RESET}        View container logs'
	@echo '  ${YELLOW}shell${RESET}       Open shell in container'
	@echo '  ${YELLOW}clean${RESET}       Remove containers and images'
	@echo '  ${YELLOW}dev${RESET}         Run development environment'
	@echo '  ${YELLOW}test${RESET}        Run tests'
	@echo ''

# Build Docker image
build:
	@echo "${GREEN}Building Docker image...${RESET}"
	docker compose build

# Run containers
run:
	@echo "${GREEN}Starting containers...${RESET}"
	docker compose up -d
	@echo "${GREEN}Home Agent is running at http://localhost:8080${RESET}"

# Stop containers
stop:
	@echo "${YELLOW}Stopping containers...${RESET}"
	docker compose down

# Restart containers
restart: stop run

# View logs
logs:
	docker compose logs -f

# Open shell in container
shell:
	docker compose exec home-agent /bin/sh

# Clean up
clean:
	@echo "${YELLOW}Cleaning up containers and images...${RESET}"
	docker compose down -v --rmi local
	@echo "${GREEN}Cleanup complete${RESET}"

# Development mode (without Docker)
dev:
	@echo "${GREEN}Starting development environment...${RESET}"
	@./start-dev.sh

# Run tests
test:
	@echo "${GREEN}Running tests...${RESET}"
	cd frontend && npm run check
	cd backend && go test ./...

# Build production
build-prod:
	@echo "${GREEN}Building for production...${RESET}"
	@./build-prod.sh
