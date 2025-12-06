#!/bin/bash
#
# Grafana Integration Test Script
# Tests real-time data flow: TFDrift-Falco -> Promtail -> Loki -> Grafana
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GRAFANA_DIR="$PROJECT_ROOT/dashboards/grafana"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

test_result() {
    local test_name="$1"
    local result="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$result" == "pass" ]; then
        log_success "✓ $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "✗ $test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

wait_for_service() {
    local service_name="$1"
    local url="$2"
    local max_attempts=30
    local attempt=1

    log_info "Waiting for $service_name to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            log_success "$service_name is ready"
            return 0
        fi

        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "$service_name did not become ready in time"
    return 1
}

# Test 1: Docker daemon check
test_docker_daemon() {
    log_info "Test 1: Checking Docker daemon..."

    if docker info > /dev/null 2>&1; then
        test_result "Docker daemon is running" "pass"
        return 0
    else
        test_result "Docker daemon is running" "fail"
        log_warning "Please start Docker Desktop and try again"
        return 1
    fi
}

# Test 2: Grafana stack startup
test_grafana_stack_startup() {
    log_info "Test 2: Starting Grafana stack..."

    cd "$GRAFANA_DIR"

    if docker-compose up -d > /dev/null 2>&1; then
        test_result "Grafana stack startup" "pass"
        return 0
    else
        test_result "Grafana stack startup" "fail"
        return 1
    fi
}

# Test 3: Service health checks
test_service_health() {
    log_info "Test 3: Checking service health..."

    # Check Loki
    if wait_for_service "Loki" "http://localhost:3100/ready"; then
        test_result "Loki health check" "pass"
    else
        test_result "Loki health check" "fail"
        return 1
    fi

    # Check Grafana
    if wait_for_service "Grafana" "http://localhost:3000/api/health"; then
        test_result "Grafana health check" "pass"
    else
        test_result "Grafana health check" "fail"
        return 1
    fi

    # Check Promtail (it doesn't have a health endpoint, so check if container is running)
    if docker-compose ps promtail | grep -q "Up"; then
        test_result "Promtail health check" "pass"
    else
        test_result "Promtail health check" "fail"
        return 1
    fi
}

# Test 4: Sample data ingestion
test_sample_data_ingestion() {
    log_info "Test 4: Verifying sample data ingestion..."

    # Wait a bit for Promtail to collect sample logs
    sleep 5

    # Query Loki for tfdrift-falco logs
    local query_result=$(curl -s "http://localhost:3100/loki/api/v1/query" \
        --data-urlencode 'query={job="tfdrift-falco"}' \
        --data-urlencode 'limit=1')

    if echo "$query_result" | jq -e '.status == "success" and (.data.result | length) > 0' > /dev/null 2>&1; then
        local event_count=$(echo "$query_result" | jq -r '.data.result | length')
        log_info "Found $event_count log streams in Loki"
        test_result "Sample data ingestion to Loki" "pass"
    else
        log_warning "No data found in Loki yet. This is expected if starting fresh."
        test_result "Sample data ingestion to Loki" "fail"
        return 1
    fi
}

# Test 5: Dashboard queries
test_dashboard_queries() {
    log_info "Test 5: Testing dashboard queries..."

    # Test query: count drift events
    local query='count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [24h])'
    local query_result=$(curl -s "http://localhost:3100/loki/api/v1/query" \
        --data-urlencode "query=$query")

    if echo "$query_result" | jq -e '.status == "success"' > /dev/null 2>&1; then
        test_result "Dashboard query execution" "pass"
    else
        test_result "Dashboard query execution" "fail"
        return 1
    fi
}

# Test 6: Grafana datasource connection
test_grafana_datasource() {
    log_info "Test 6: Checking Grafana datasource configuration..."

    # Login to Grafana and get datasources
    local datasources=$(curl -s -u admin:admin "http://localhost:3000/api/datasources")

    if echo "$datasources" | jq -e '.[] | select(.type == "loki")' > /dev/null 2>&1; then
        local ds_name=$(echo "$datasources" | jq -r '.[] | select(.type == "loki") | .name')
        log_info "Loki datasource found: $ds_name"
        test_result "Grafana datasource configuration" "pass"
    else
        test_result "Grafana datasource configuration" "fail"
        return 1
    fi
}

# Test 7: Dashboard provisioning
test_dashboard_provisioning() {
    log_info "Test 7: Checking dashboard provisioning..."

    # Get list of dashboards
    local dashboards=$(curl -s -u admin:admin "http://localhost:3000/api/search?type=dash-db")

    local expected_dashboards=("TFDrift-Falco Overview" "TFDrift-Falco Diff Details" "TFDrift-Falco Heatmap")
    local found_count=0

    for dashboard in "${expected_dashboards[@]}"; do
        if echo "$dashboards" | jq -e ".[] | select(.title == \"$dashboard\")" > /dev/null 2>&1; then
            log_info "Dashboard found: $dashboard"
            found_count=$((found_count + 1))
        else
            log_warning "Dashboard not found: $dashboard"
        fi
    done

    if [ $found_count -eq ${#expected_dashboards[@]} ]; then
        test_result "Dashboard provisioning (all 3 dashboards)" "pass"
    else
        test_result "Dashboard provisioning ($found_count/3 dashboards)" "fail"
        return 1
    fi
}

# Test 8: Real-time log generation (optional)
test_realtime_log_generation() {
    log_info "Test 8: Testing real-time log generation..."

    # Generate a test drift event
    local test_log_file="$GRAFANA_DIR/sample-logs/test-$(date +%s).jsonl"
    local test_event=$(cat <<EOF
{"timestamp":"$(date -u +%Y-%m-%dT%H:%M:%SZ)","resource_type":"aws_security_group","resource_id":"sg-test-$(date +%s)","changed_by":"integration-test","severity":"high","diff":{"ingress":{"expected":["443/tcp"],"actual":["443/tcp","22/tcp"]}},"action":"drift_detected"}
EOF
)

    echo "$test_event" > "$test_log_file"
    log_info "Generated test event: $test_log_file"

    # Wait for Promtail to pick it up
    sleep 5

    # Query Loki for the test event
    local resource_id=$(echo "$test_event" | jq -r '.resource_id')
    local query_result=$(curl -s "http://localhost:3100/loki/api/v1/query" \
        --data-urlencode "query={job=\"tfdrift-falco\"} | json | resource_id=\"$resource_id\"")

    if echo "$query_result" | jq -e '.status == "success" and (.data.result | length) > 0' > /dev/null 2>&1; then
        log_info "Test event successfully ingested and queryable"
        test_result "Real-time log generation and ingestion" "pass"
        rm -f "$test_log_file"
    else
        log_warning "Test event not found in Loki (may need more time)"
        test_result "Real-time log generation and ingestion" "fail"
        return 1
    fi
}

# Test 9: Performance check
test_performance() {
    log_info "Test 9: Checking query performance..."

    local start_time=$(date +%s%3N)
    curl -s "http://localhost:3100/loki/api/v1/query" \
        --data-urlencode 'query={job="tfdrift-falco"} | json | action="drift_detected"' \
        --data-urlencode 'limit=100' > /dev/null
    local end_time=$(date +%s%3N)

    local duration=$((end_time - start_time))
    log_info "Query execution time: ${duration}ms"

    if [ $duration -lt 5000 ]; then
        test_result "Query performance (<5s)" "pass"
    else
        test_result "Query performance (${duration}ms)" "fail"
        return 1
    fi
}

# Main test execution
main() {
    echo "=============================================="
    echo "  TFDrift-Falco Grafana Integration Test"
    echo "=============================================="
    echo ""

    # Run tests
    test_docker_daemon || {
        log_error "Docker is not running. Please start Docker Desktop first."
        exit 1
    }

    test_grafana_stack_startup
    sleep 5  # Give services time to initialize

    test_service_health
    test_sample_data_ingestion
    test_dashboard_queries
    test_grafana_datasource
    test_dashboard_provisioning
    test_realtime_log_generation
    test_performance

    # Summary
    echo ""
    echo "=============================================="
    echo "  Test Summary"
    echo "=============================================="
    echo "Total Tests:  $TOTAL_TESTS"
    echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "All tests passed! ✓"
        echo ""
        echo "Access Grafana at: http://localhost:3000"
        echo "Username: admin"
        echo "Password: admin"
        echo ""
        exit 0
    else
        log_error "Some tests failed. Please check the output above."
        echo ""
        echo "To view logs:"
        echo "  cd $GRAFANA_DIR"
        echo "  docker-compose logs grafana"
        echo "  docker-compose logs loki"
        echo "  docker-compose logs promtail"
        echo ""
        exit 1
    fi
}

# Cleanup function (optional)
cleanup() {
    log_info "Cleaning up test environment..."
    cd "$GRAFANA_DIR"
    docker-compose down > /dev/null 2>&1 || true
    log_success "Cleanup complete"
}

# Parse command line arguments
case "${1:-}" in
    cleanup)
        cleanup
        exit 0
        ;;
    --help|-h)
        echo "Usage: $0 [cleanup|--help]"
        echo ""
        echo "Run Grafana integration tests for TFDrift-Falco"
        echo ""
        echo "Options:"
        echo "  cleanup    Stop and remove all containers"
        echo "  --help     Show this help message"
        exit 0
        ;;
    *)
        main
        ;;
esac
