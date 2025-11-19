# Grafana Integration (Preview)

This directory contains the initial setup to visualize tfdrift-falco drift events using Grafana, Loki, and Promtail.

The goal is to provide an out-of-the-box dashboard that allows users to:

- View drift events visually
- Filter by resource type, severity, and actor
- Compare expected vs actual values
- Monitor drift activity over time

## What's Included

This PR includes:

- A minimal docker-compose stack (Grafana + Loki + Promtail)
- A sample drift JSON log
- Promtail configuration for indexing drift logs

## Getting Started

To try the preview:

```bash
cd dashboards/grafana
docker-compose up -d
```

Then open Grafana at: http://localhost:3000

- **Username**: admin
- **Password**: admin

## Next Steps

Future PRs will add:

- Full Grafana dashboard JSON
- Expected vs Actual diff panels
- Drift heatmaps and timelines
- Real-time alerting rules
- Visualization examples and screenshots
