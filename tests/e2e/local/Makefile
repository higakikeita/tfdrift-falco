.PHONY: help test up down logs clean build build-mock

# Default target
help:
	@echo "TFDrift-Falco Local E2E Test Environment"
	@echo ""
	@echo "Available targets:"
	@echo "  make up              - Start all services (docker-compose up -d)"
	@echo "  make down            - Stop and remove all services"
	@echo "  make logs            - Follow logs from all services"
	@echo "  make logs-falco      - Follow logs from mock Falco server"
	@echo "  make logs-tfdrift    - Follow logs from TFDrift API server"
	@echo "  make test            - Run E2E tests (requires services running)"
	@echo "  make test-quick      - Quick test run (no service startup)"
	@echo "  make clean           - Remove all containers and volumes"
	@echo "  make build           - Build Docker images"
	@echo "  make build-mock      - Build mock Falco server binary locally"
	@echo "  make ps              - Show running containers"
	@echo "  make shell-falco     - Open shell in mock Falco container"
	@echo "  make shell-tfdrift   - Open shell in TFDrift container"
	@echo ""
	@echo "Example workflow:"
	@echo "  1. make build       # Build images"
	@echo "  2. make up          # Start services"
	@echo "  3. make logs        # Watch logs in another terminal"
	@echo "  4. make test        # Run tests"
	@echo "  5. make down        # Stop services"

# Start all services
up:
	docker-compose -f docker-compose.yml up -d
	@echo "Services started. Run 'make logs' to follow logs."

# Stop all services
down:
	docker-compose -f docker-compose.yml down

# Follow all logs
logs:
	docker-compose -f docker-compose.yml logs -f

# Follow Falco logs only
logs-falco:
	docker-compose -f docker-compose.yml logs -f mock-falco

# Follow TFDrift logs only
logs-tfdrift:
	docker-compose -f docker-compose.yml logs -f tfdrift

# Run E2E tests (assumes services are running)
test:
	cd ../.. && go test -v -tags e2e -count=1 ./tests/e2e/local/...

# Quick test without rebuilding everything
test-quick:
	cd ../.. && go test -v -tags e2e -count=1 -short ./tests/e2e/local/...

# Build Docker images
build:
	docker-compose -f docker-compose.yml build

# Build mock Falco server binary locally
build-mock:
	cd ../.. && go build -o tests/e2e/local/mock-falco-server ./tests/e2e/local/mock_falco_server.go

# Show running containers
ps:
	docker-compose -f docker-compose.yml ps

# Clean up all containers and volumes
clean: down
	docker-compose -f docker-compose.yml rm -f
	docker volume rm tfdrift_data tfdrift_config 2>/dev/null || true

# Open shell in mock Falco container
shell-falco:
	docker-compose -f docker-compose.yml exec mock-falco sh

# Open shell in TFDrift container
shell-tfdrift:
	docker-compose -f docker-compose.yml exec tfdrift sh

# Full test cycle: build -> up -> test -> down
full-test: build up test down

# Health check
health:
	@echo "Checking Falco health..."
	@curl -s http://localhost:8081/health | jq .
	@echo ""
	@echo "Checking TFDrift health..."
	@curl -s http://localhost:8080/health | jq .
