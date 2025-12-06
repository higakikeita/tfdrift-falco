#!/bin/bash

# TFDrift-Falco Load Test Runner
# This script orchestrates the complete load testing process

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Test scenarios
declare -A SCENARIOS=(
    ["small"]="100:500:1h"
    ["medium"]="1000:5000:4h"
    ["large"]="10000:50000:8h"
)

usage() {
    cat << EOF
Usage: $0 [OPTIONS] <scenario>

Run TFDrift-Falco load tests with specified scenario.

Scenarios:
  small   - 100 events/min, 500 resources, 1 hour
  medium  - 1000 events/min, 5000 resources, 4 hours
  large   - 10000 events/min, 50000 resources, 8 hours

Options:
  -h, --help              Show this help message
  -o, --output DIR        Output directory (default: /tmp/tfdrift-load-test)
  -c, --cleanup           Cleanup after test
  --skip-build            Skip Docker image build
  --skip-generate         Skip data generation (use existing)

Examples:
  $0 small
  $0 -o ./results medium
  $0 --cleanup large

EOF
    exit 0
}

# Parse arguments
SCENARIO=""
OUTPUT_DIR="/tmp/tfdrift-load-test"
CLEANUP=false
SKIP_BUILD=false
SKIP_GENERATE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -c|--cleanup)
            CLEANUP=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-generate)
            SKIP_GENERATE=true
            shift
            ;;
        small|medium|large)
            SCENARIO="$1"
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

if [ -z "$SCENARIO" ]; then
    log_error "No scenario specified"
    usage
fi

# Parse scenario parameters
IFS=':' read -r EVENT_RATE RESOURCES DURATION <<< "${SCENARIOS[$SCENARIO]}"

log_info "TFDrift-Falco Load Test"
log_info "========================"
log_info "Scenario: $SCENARIO"
log_info "  Event Rate: $EVENT_RATE events/min"
log_info "  Resources: $RESOURCES"
log_info "  Duration: $DURATION"
log_info "  Output: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Step 1: Generate test data
if [ "$SKIP_GENERATE" = false ]; then
    log_step "Step 1: Generating test data"

    # Generate CloudTrail events
    log_info "Generating CloudTrail events..."
    go run cloudtrail_simulator.go \
        --rate "$EVENT_RATE" \
        --duration "$DURATION" \
        --output "$OUTPUT_DIR/simulated-cloudtrail-logs"

    # Generate Terraform state
    log_info "Generating Terraform state..."
    go run terraform_state_generator.go \
        --resources "$RESOURCES" \
        --output "$OUTPUT_DIR/generated-terraform-state/terraform.tfstate"

    log_info "Test data generation complete"
    echo ""
else
    log_info "Skipping data generation (--skip-generate)"
fi

# Step 2: Build Docker image
if [ "$SKIP_BUILD" = false ]; then
    log_step "Step 2: Building Docker image"
    docker build -t tfdrift-falco:load-test ../../
    log_info "Docker build complete"
    echo ""
else
    log_info "Skipping Docker build (--skip-build)"
fi

# Step 3: Create configuration
log_step "Step 3: Creating configuration"

cat > "$OUTPUT_DIR/config.yaml" << EOF
# TFDrift-Falco Load Test Configuration

providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: /terraform/terraform.tfstate

falco:
  enabled: true
  hostname: falco
  port: 5060
  tls: false

drift_rules:
  - name: "All Resources"
    resource_types:
      - "aws_instance"
      - "aws_iam_role"
      - "aws_s3_bucket"
      - "aws_db_instance"
      - "aws_lambda_function"
    watched_attributes:
      - "*"
    severity: "high"

logging:
  level: "info"
  format: "json"
  output: "file"
  file: "/var/log/tfdrift/tfdrift.jsonl"

advanced:
  state_refresh_interval: "5m"
  workers: 8
  event_buffer_size: 1000
EOF

log_info "Configuration created"
echo ""

# Step 4: Start test environment
log_step "Step 4: Starting test environment"

cd "$(dirname "$0")"

# Copy generated data to docker-compose location
ln -sf "$OUTPUT_DIR/simulated-cloudtrail-logs" ./simulated-cloudtrail-logs
ln -sf "$OUTPUT_DIR/generated-terraform-state" ./generated-terraform-state
ln -sf "$OUTPUT_DIR/config.yaml" ./config.yaml

docker-compose -f docker-compose.load-test.yml up -d

log_info "Waiting for services to be ready..."
sleep 30

# Check service health
log_info "Checking service health..."
docker-compose -f docker-compose.load-test.yml ps

echo ""

# Step 5: Collect metrics
log_step "Step 5: Collecting metrics"

# Convert duration to seconds
DURATION_SECONDS=$(echo "$DURATION" | sed 's/h/*3600/g' | sed 's/m/*60/g' | sed 's/s//g' | bc)

log_info "Starting metrics collection for ${DURATION_SECONDS}s..."

export OUTPUT_DIR="$OUTPUT_DIR/metrics"
export INTERVAL=10
export DURATION="$DURATION_SECONDS"

./collect_metrics.sh

log_info "Metrics collection complete"
echo ""

# Step 6: Analyze results
log_step "Step 6: Analyzing results"

if command -v python3 &> /dev/null; then
    log_info "Running analysis script..."
    # TODO: Implement analysis script
    log_warn "Analysis script not yet implemented"
else
    log_warn "Python3 not found, skipping analysis"
fi

echo ""

# Step 7: Generate report
log_step "Step 7: Generating report"

cat > "$OUTPUT_DIR/report.md" << EOF
# TFDrift-Falco Load Test Report

**Scenario**: $SCENARIO
**Date**: $(date)

## Test Configuration

- **Event Rate**: $EVENT_RATE events/min
- **Resources**: $RESOURCES Terraform resources
- **Duration**: $DURATION
- **Output**: $OUTPUT_DIR

## Results

See detailed metrics in:
- \`metrics/summary.txt\` - Overall summary
- \`metrics/cpu_memory.csv\` - CPU and memory usage
- \`metrics/docker_stats.csv\` - Docker container stats
- \`metrics/prometheus_metrics.txt\` - Prometheus metrics
- \`metrics/loki_queries.txt\` - Loki event counts

## Grafana Dashboard

Access Grafana at: http://localhost:3000
- Username: admin
- Password: admin

## Next Steps

1. Review metrics in Grafana
2. Check for errors in logs: \`docker-compose -f docker-compose.load-test.yml logs\`
3. Analyze performance bottlenecks
4. Tune configuration if needed

EOF

log_info "Report generated: $OUTPUT_DIR/report.md"
cat "$OUTPUT_DIR/report.md"

echo ""

# Step 8: Cleanup (optional)
if [ "$CLEANUP" = true ]; then
    log_step "Step 8: Cleanup"
    log_warn "Stopping and removing containers..."
    docker-compose -f docker-compose.load-test.yml down -v
    log_info "Cleanup complete"
else
    log_info "Test environment is still running. To stop:"
    echo "  docker-compose -f docker-compose.load-test.yml down"
fi

echo ""
log_info "Load test complete! 🎉"
log_info "Results: $OUTPUT_DIR"
