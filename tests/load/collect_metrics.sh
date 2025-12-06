#!/bin/bash

# Metrics Collection Script for TFDrift-Falco Load Testing
# This script collects performance metrics during load tests

set -euo pipefail

# Configuration
OUTPUT_DIR="${OUTPUT_DIR:-/tmp/tfdrift-load-test-metrics}"
INTERVAL="${INTERVAL:-5}"  # seconds
DURATION="${DURATION:-3600}"  # seconds (1 hour default)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Create output directory
mkdir -p "$OUTPUT_DIR"

log_info "Starting metrics collection"
log_info "  Output directory: $OUTPUT_DIR"
log_info "  Interval: ${INTERVAL}s"
log_info "  Duration: ${DURATION}s"

# File paths
CPU_MEMORY_FILE="$OUTPUT_DIR/cpu_memory.csv"
DOCKER_STATS_FILE="$OUTPUT_DIR/docker_stats.csv"
PROMETHEUS_METRICS_FILE="$OUTPUT_DIR/prometheus_metrics.txt"
LOKI_QUERIES_FILE="$OUTPUT_DIR/loki_queries.txt"
SUMMARY_FILE="$OUTPUT_DIR/summary.txt"

# Initialize CSV files
echo "timestamp,cpu_percent,memory_mb,memory_percent" > "$CPU_MEMORY_FILE"
echo "timestamp,container,cpu_percent,memory_usage,memory_limit,memory_percent,net_io,block_io" > "$DOCKER_STATS_FILE"

# Trap to handle cleanup
trap cleanup EXIT INT TERM

cleanup() {
    log_info "Stopping metrics collection..."
    kill $DOCKER_PID $PROM_PID 2>/dev/null || true
    generate_summary
    log_info "Metrics saved to: $OUTPUT_DIR"
}

# Function to collect docker stats
collect_docker_stats() {
    while true; do
        timestamp=$(date +%s)
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}" | \
        tail -n +2 | \
        while IFS=$'\t' read -r name cpu mem_usage mem_perc net_io block_io; do
            # Extract memory usage and limit
            mem_used=$(echo "$mem_usage" | awk '{print $1}')
            mem_limit=$(echo "$mem_usage" | awk '{print $3}')

            echo "$timestamp,$name,$cpu,$mem_used,$mem_limit,$mem_perc,$net_io,$block_io" >> "$DOCKER_STATS_FILE"
        done
        sleep "$INTERVAL"
    done
}

# Function to collect Prometheus metrics
collect_prometheus_metrics() {
    while true; do
        timestamp=$(date +%s)
        echo "=== Timestamp: $timestamp ===" >> "$PROMETHEUS_METRICS_FILE"

        # TFDrift-Falco metrics
        if curl -s http://localhost:9090/metrics > /dev/null 2>&1; then
            curl -s http://localhost:9090/metrics | grep -E "^tfdrift_" >> "$PROMETHEUS_METRICS_FILE" 2>/dev/null || true
        fi

        echo "" >> "$PROMETHEUS_METRICS_FILE"
        sleep "$INTERVAL"
    done
}

# Function to query Loki for event counts
collect_loki_queries() {
    while true; do
        timestamp=$(date +%s)
        echo "=== Timestamp: $timestamp ===" >> "$LOKI_QUERIES_FILE"

        # Total events in last 5 minutes
        if command -v logcli &> /dev/null; then
            logcli query --limit=0 --since=5m '{job="tfdrift-falco"}' --stats 2>/dev/null >> "$LOKI_QUERIES_FILE" || true
        elif command -v curl &> /dev/null; then
            # Fallback to curl
            curl -s -G "http://localhost:3100/loki/api/v1/query_range" \
                --data-urlencode 'query={job="tfdrift-falco"}' \
                --data-urlencode "start=$(date -u -d '5 minutes ago' +%s)000000000" \
                --data-urlencode "end=$(date -u +%s)000000000" \
                | jq -r '.data.result[] | .values | length' >> "$LOKI_QUERIES_FILE" 2>/dev/null || true
        fi

        echo "" >> "$LOKI_QUERIES_FILE"
        sleep "$INTERVAL"
    done
}

# Function to generate summary
generate_summary() {
    log_info "Generating summary..."

    {
        echo "======================================"
        echo "TFDrift-Falco Load Test Summary"
        echo "======================================"
        echo ""
        echo "Test Duration: ${DURATION}s"
        echo "Collection Interval: ${INTERVAL}s"
        echo "Generated at: $(date)"
        echo ""

        echo "Docker Container Stats:"
        echo "----------------------"
        if [ -f "$DOCKER_STATS_FILE" ]; then
            # Calculate averages for TFDrift container
            echo "TFDrift-Falco Container:"
            awk -F',' '/tfdrift/ {
                cpu_sum += $3;
                mem_sum += $4;
                count++
            }
            END {
                if (count > 0) {
                    printf "  Average CPU: %.2f%%\n", cpu_sum/count;
                    printf "  Average Memory: %.2f MB\n", mem_sum/count;
                    printf "  Samples: %d\n", count
                }
            }' "$DOCKER_STATS_FILE"

            echo ""
            echo "Falco Container:"
            awk -F',' '/falco/ {
                cpu_sum += $3;
                mem_sum += $4;
                count++
            }
            END {
                if (count > 0) {
                    printf "  Average CPU: %.2f%%\n", cpu_sum/count;
                    printf "  Average Memory: %.2f MB\n", mem_sum/count;
                    printf "  Samples: %d\n", count
                }
            }' "$DOCKER_STATS_FILE"
        fi

        echo ""
        echo "Prometheus Metrics:"
        echo "-------------------"
        if [ -f "$PROMETHEUS_METRICS_FILE" ]; then
            # Extract key metrics
            echo "Event Processing:"
            grep -E "tfdrift_events_processed_total" "$PROMETHEUS_METRICS_FILE" | tail -1 || echo "  N/A"

            echo ""
            echo "Drift Alerts:"
            grep -E "tfdrift_drift_alerts_total" "$PROMETHEUS_METRICS_FILE" | tail -1 || echo "  N/A"

            echo ""
            echo "Processing Time (p95):"
            grep -E "tfdrift_event_processing_duration.*0.95" "$PROMETHEUS_METRICS_FILE" | tail -1 || echo "  N/A"
        fi

        echo ""
        echo "Loki Event Counts:"
        echo "------------------"
        if [ -f "$LOKI_QUERIES_FILE" ]; then
            total_events=$(grep -E "^[0-9]+$" "$LOKI_QUERIES_FILE" | awk '{sum += $1} END {print sum}')
            echo "  Total Events Logged: ${total_events:-0}"
        fi

        echo ""
        echo "======================================"
        echo "Detailed data available in:"
        echo "  CPU/Memory: $CPU_MEMORY_FILE"
        echo "  Docker Stats: $DOCKER_STATS_FILE"
        echo "  Prometheus: $PROMETHEUS_METRICS_FILE"
        echo "  Loki: $LOKI_QUERIES_FILE"
        echo "======================================"
    } | tee "$SUMMARY_FILE"
}

# Start background collectors
log_info "Starting background metric collectors..."

collect_docker_stats &
DOCKER_PID=$!

collect_prometheus_metrics &
PROM_PID=$!

# Wait for duration
log_info "Collecting metrics for ${DURATION}s..."
sleep "$DURATION"

log_info "Collection complete"
