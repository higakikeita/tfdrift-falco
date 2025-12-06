# Grafana Integration

This directory contains a complete Grafana observability stack for visualizing tfdrift-falco drift events using Grafana, Loki, and Promtail.

## Quick Links

- 🚀 **[Getting Started Guide](./GETTING_STARTED.md)** - **START HERE** - Complete setup guide for end users
- 📖 **[Integration Test Results](./INTEGRATION_TEST_RESULTS.md)** - Detailed test report and verification
- 🚨 **[Alert Configuration Guide](./ALERTS.md)** - Step-by-step alert setup instructions
- 🎨 **[Customization Guide](./CUSTOMIZATION_GUIDE.md)** - Dashboard customization and best practices
- 🧪 **[Integration Test Script](../../tests/integration/test_grafana.sh)** - Automated testing

## Features

### 📊 Three Pre-Built Dashboards

1. **TFDrift-Falco Overview** - Main dashboard with:
   - Total drift events counter
   - Drift events by severity (pie chart)
   - Drift events by resource type (donut chart)
   - Unique resources and actors affected
   - Timeline view with severity breakdown
   - Recent drift events log panel
   - Detailed drift table with filtering

2. **TFDrift-Falco Diff Details** - Deep-dive into configuration changes:
   - Side-by-side expected vs actual value comparison
   - JSON diff viewer for complex objects
   - Changes by actor breakdown
   - Top 10 resources with most drift
   - Resource-specific filtering

3. **TFDrift-Falco Heatmap & Analytics** - Pattern analysis:
   - Drift frequency heatmap over time
   - Activity by resource type (bar chart)
   - Activity by severity level (bar chart)
   - Resource type × severity matrix
   - Hourly drift trends

### 🔍 Key Capabilities

- **Real-time monitoring** with 5-30s auto-refresh
- **Multi-dimensional filtering** by severity, resource type, resource ID
- **Color-coded severity levels**: Critical (dark red), High (red), Medium (orange), Low (yellow)
- **Auto-provisioning** of datasources and dashboards
- **Sample data included** for immediate testing

## What's Included

```
dashboards/grafana/
├── docker-compose.yaml           # Full stack orchestration
├── provisioning/
│   ├── datasources/
│   │   └── loki.yaml            # Loki datasource config
│   └── dashboards/
│       └── default.yaml         # Dashboard auto-loading config
├── dashboards/
│   ├── tfdrift-overview.json    # Main overview dashboard
│   ├── tfdrift-diff-details.json # Diff comparison dashboard
│   └── tfdrift-heatmap.json     # Analytics & heatmap dashboard
├── sample-logs/
│   ├── drift-events.jsonl       # 20 sample drift events
│   └── sample-drift.json        # Original sample event
└── promtail-config.yaml         # Log collection config
```

## Getting Started

### Quick Start (5 minutes)

```bash
# 1. Start the Grafana stack
cd dashboards/grafana
docker-compose up -d

# 2. Open your browser
open http://localhost:3000

# 3. Login (username: admin, password: admin)

# 4. Navigate to Dashboards → TFDrift-Falco folder
```

**See sample data immediately!** All dashboards are pre-loaded with 20+ sample drift events.

### Connect to Real Data

To monitor actual TFDrift-Falco drift events, see the **[Getting Started Guide](./GETTING_STARTED.md)** for:
- Connecting TFDrift-Falco logs
- Setting up alerts
- Real-world usage examples
- Troubleshooting

**TL;DR**: Mount your TFDrift-Falco log directory to Promtail and restart.

## Using with Real Data

To integrate with your tfdrift-falco deployment:

1. Configure tfdrift-falco to output JSON logs:
   ```yaml
   output:
     format: json
     file: /var/log/tfdrift/drift-events.jsonl
   ```

2. Update `promtail-config.yaml` to point to your log directory:
   ```yaml
   scrape_configs:
     - job_name: "tfdrift-falco"
       static_configs:
         - targets: [localhost]
           labels:
             job: tfdrift-falco
             __path__: /path/to/your/logs/*.jsonl
   ```

3. Mount your log directory in `docker-compose.yaml`:
   ```yaml
   promtail:
     volumes:
       - /path/to/your/logs:/var/log/tfdrift
   ```

4. Restart the stack:
   ```bash
   docker-compose restart
   ```

## Dashboard Queries

All dashboards use LogQL (Loki Query Language). Example queries:

```logql
# Count drift events
count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [5m])

# Group by severity
sum by (severity) (count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [1h]))

# Filter by resource type
{job="tfdrift-falco"} | json | action="drift_detected" | resource_type="aws_security_group"
```

## Customization

All dashboard JSON files are editable. Modify them to:
- Add custom panels
- Change time ranges
- Adjust severity thresholds
- Create custom alerts

After editing, restart Grafana:
```bash
docker-compose restart grafana
```

## Troubleshooting

### No data showing in dashboards

```bash
# Check if Promtail is reading logs
docker logs grafana-promtail-1

# Check if Loki is receiving data
curl -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={job="tfdrift-falco"}' | jq

# Verify log files are mounted
docker exec grafana-promtail-1 ls -la /var/log/tfdrift/
```

### Dashboards not loading

```bash
# Check Grafana logs
docker logs grafana-grafana-1

# Verify dashboard files
docker exec grafana-grafana-1 ls -la /var/lib/grafana/dashboards/
```

## Cleanup

```bash
docker-compose down
```

To remove all data volumes:
```bash
docker-compose down -v
```
