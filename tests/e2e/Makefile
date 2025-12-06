.PHONY: help setup test cleanup all test-quick test-full

## help: Show this help message
help:
	@echo "TFDrift-Falco E2E Tests"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /' | column -t -s ':'

## setup: Set up test infrastructure (Terraform apply)
setup:
	@echo "Setting up E2E test infrastructure..."
	cd terraform && terraform init -backend-config=backend.hcl || terraform init
	cd terraform && terraform apply -auto-approve
	cd terraform && terraform output -json > ../fixtures/terraform_outputs.json
	@echo "✅ Infrastructure ready"

## test: Run E2E tests (requires setup first)
test:
	@echo "Running E2E tests..."
	@echo "⚠️  These tests take 20-30 minutes due to CloudTrail delays"
	go test -v -timeout 45m ./...

## test-quick: Run quick E2E tests (no CloudTrail delay)
test-quick:
	@echo "Running quick E2E tests..."
	go test -v -short ./...

## test-full: Setup + Test + Cleanup
test-full: setup test cleanup

## cleanup: Destroy test infrastructure
cleanup:
	@echo "Cleaning up E2E test infrastructure..."
	cd terraform && terraform destroy -auto-approve
	@echo "✅ Cleanup complete"

## docker-up: Start Falco with Docker Compose
docker-up:
	@echo "Starting Falco..."
	docker-compose -f docker-compose.e2e.yml up -d
	@echo "Waiting for Falco to be ready..."
	sleep 5
	@echo "✅ Falco ready"

## docker-down: Stop Falco
docker-down:
	@echo "Stopping Falco..."
	docker-compose -f docker-compose.e2e.yml down
	@echo "✅ Falco stopped"

## docker-logs: View Falco logs
docker-logs:
	docker-compose -f docker-compose.e2e.yml logs -f falco

## verify: Verify prerequisites
verify:
	@echo "Verifying prerequisites..."
	@command -v terraform >/dev/null 2>&1 || { echo "❌ terraform not found"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "❌ docker not found"; exit 1; }
	@command -v aws >/dev/null 2>&1 || { echo "❌ aws cli not found"; exit 1; }
	@aws sts get-caller-identity >/dev/null 2>&1 || { echo "❌ AWS credentials not configured"; exit 1; }
	@echo "✅ All prerequisites met"

## all: Complete E2E test cycle (verify + docker + setup + test + cleanup)
all: verify docker-up setup test cleanup docker-down

.DEFAULT_GOAL := help
