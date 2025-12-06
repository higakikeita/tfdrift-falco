# TFDrift-Falco Grafana Customization Guide

This guide shows you how to customize the TFDrift-Falco Grafana dashboards for your specific needs.

## Table of Contents

1. [Adding Custom Panels](#adding-custom-panels)
2. [Creating Custom Queries](#creating-custom-queries)
3. [Modifying Visualizations](#modifying-visualizations)
4. [Dashboard Variables](#dashboard-variables)
5. [Color Schemes and Themes](#color-schemes-and-themes)
6. [Exporting and Sharing](#exporting-and-sharing)
7. [Common Customizations](#common-customizations)

---

## Adding Custom Panels

### Example 1: Add "Drifts by Actor" Panel

1. **Open Dashboard**
   - Navigate to **TFDrift-Falco Overview**
   - Click **Edit** (pencil icon) in top-right

2. **Add Panel**
   - Click **+ Add** → **Visualization**

3. **Configure Query**
   - Data source: `Loki`
   - Query:
     ```logql
     sum by (changed_by) (count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [$__range]))
     ```

4. **Choose Visualization**
   - Panel type: **Bar chart**
   - Orientation: **Horizontal**

5. **Panel Settings**
   - Title: `Drift Events by Actor`
   - Description: `Shows who made the most configuration changes`

6. **Save**
   - Click **Apply** → **Save dashboard**

### Example 2: Add "Critical Events Table"

1. **Add Visualization**
   - Panel type: **Table**

2. **Query**
   ```logql
   {job="tfdrift-falco"} | json | severity="critical" | line_format "{{.timestamp}} | {{.resource_type}} | {{.resource_id}} | {{.changed_by}}"
   ```

3. **Transform Data**
   - Add transformation: **Extract fields**
   - Format: `\s*\|\s*` (pipe-delimited)
   - Fields: timestamp, resource_type, resource_id, changed_by

4. **Table Settings**
   - Column alignment: Left
   - Enable sorting
   - Add column links (optional)

---

## Creating Custom Queries

### Query Patterns

#### 1. Time-based Aggregation

```logql
# Drifts per hour
sum by (hour) (count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [1h]))

# Drifts per day
sum(count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [24h]))
```

#### 2. Multi-condition Filtering

```logql
# High or Critical severity only
{job="tfdrift-falco"} | json | severity=~"high|critical"

# Security-related resources
{job="tfdrift-falco"} | json | resource_type=~"aws_security_group|aws_iam_.*|aws_kms_.*"

# Exclude specific actors
{job="tfdrift-falco"} | json | changed_by!~"terraform|automation"
```

#### 3. Complex Aggregations

```logql
# Average drifts per resource type
avg by (resource_type) (count_over_time({job="tfdrift-falco"} | json [$__range]))

# Top 5 most drifted resources
topk(5, sum by (resource_id) (count_over_time({job="tfdrift-falco"} | json [$__range])))

# Percentage by severity
sum by (severity) (count_over_time({job="tfdrift-falco"} | json [$__range])) / ignoring(severity) group_left sum(count_over_time({job="tfdrift-falco"} | json [$__range]))
```

#### 4. Diff-based Queries

```logql
# S3 public access changes
{job="tfdrift-falco"} | json | resource_type="aws_s3_bucket" | line_match_regex "public_access_block"

# Security group ingress changes
{job="tfdrift-falco"} | json | resource_type="aws_security_group" | line_match_regex "ingress.*22/tcp|3389/tcp"

# IAM wildcard permissions
{job="tfdrift-falco"} | json | resource_type=~"aws_iam_.*" | line_match_regex "\\*"
```

### Query Builder Tips

1. **Start Simple**: Begin with `{job="tfdrift-falco"} | json`
2. **Add Filters Gradually**: Add one filter at a time and test
3. **Use Explore View**: Test queries in **Explore** before adding to dashboards
4. **Check Performance**: Use **Query inspector** to see execution time
5. **Use Variables**: Make queries reusable with dashboard variables

---

## Modifying Visualizations

### Change Panel Type

1. **Edit Panel**
2. **Click Visualization Type** dropdown (top-right)
3. **Select New Type**:
   - Time series → Line graph with time axis
   - Stat → Single number with sparkline
   - Table → Tabular data
   - Bar chart → Horizontal/vertical bars
   - Pie chart → Circular percentage
   - Heatmap → Color-coded grid
   - Gauge → Semicircular meter

### Customize Colors

#### Option 1: Threshold-based Colors

```
Panel → Field → Thresholds
- Base: green (0)
- Warning: yellow (5)
- Critical: red (10)
```

#### Option 2: Value-based Colors

```
Panel → Field → Overrides
- Add override: "Fields with name: critical"
- Color: Dark red (#8b0000)
```

#### Option 3: Gradient Colors

```
Panel → Field → Color scheme
- Select: "Green-Yellow-Red"
- Mode: Continuous gradient
```

### Panel Layout

```
Grid Position:
- X: 0-24 (left to right)
- Y: 0-infinite (top to bottom)
- W: 1-24 (width in grid units)
- H: 1-infinite (height in grid units)

Example:
- Small stat: W=4, H=3
- Medium chart: W=12, H=8
- Full-width table: W=24, H=10
```

---

## Dashboard Variables

Variables make dashboards dynamic and reusable.

### Create a Severity Filter

1. **Dashboard Settings** (gear icon)
2. **Variables** → **+ Add variable**

```
Variable Configuration:
- Name: severity_filter
- Type: Query
- Data source: Loki
- Label: Severity
- Query: label_values(severity)
- Multi-value: true
- Include All option: true
```

3. **Use in Query**

```logql
# Before
{job="tfdrift-falco"} | json | severity="critical"

# After (with variable)
{job="tfdrift-falco"} | json | severity=~"$severity_filter"
```

### Create a Resource Type Filter

```
Variable Configuration:
- Name: resource_type
- Type: Query
- Query: label_values(resource_type)
- Multi-value: true
- Include All option: true
```

### Create a Time Range Variable

```
Variable Configuration:
- Name: time_range
- Type: Custom
- Values: 5m,15m,1h,6h,24h,7d
- Default: 1h
```

Use in query:
```logql
count_over_time({job="tfdrift-falco"} | json [$time_range])
```

### Create a Refresh Interval Variable

```
Variable Configuration:
- Name: refresh_interval
- Type: Interval
- Values: 5s,10s,30s,1m,5m
- Auto: true
```

Set in dashboard settings → **Auto refresh** → `$refresh_interval`

---

## Color Schemes and Themes

### Severity Color Palette

```
Critical: #8b0000 (Dark Red)
High:     #ff0000 (Red)
Medium:   #ffa500 (Orange)
Low:      #ffff00 (Yellow)
Info:     #87ceeb (Sky Blue)
Success:  #00ff00 (Green)
```

### Custom Theme

1. **Dashboard Settings** → **JSON Model**
2. Add custom theme:

```json
{
  "style": "dark",
  "theme": {
    "colors": {
      "primary": "#ff6b6b",
      "secondary": "#4ecdc4",
      "background": "#1a1a2e",
      "text": "#eaeaea"
    }
  }
}
```

### Panel-specific Colors

```
Panel → Field → Standard options → Color scheme
- Classic palette (default)
- Green-Yellow-Red
- Blue-Yellow-Red
- Custom (define your own)
```

---

## Exporting and Sharing

### Export Dashboard JSON

1. **Dashboard Settings** → **JSON Model**
2. **Copy to Clipboard**
3. Save to file: `tfdrift-custom.json`

### Share Dashboard

```bash
# Method 1: Copy JSON file
cp dashboards/tfdrift-overview.json dashboards/tfdrift-custom.json

# Method 2: Export from UI
# Dashboard → Share → Export → Save to file

# Method 3: API export
curl -u admin:admin \
  http://localhost:3000/api/dashboards/uid/tfdrift-overview \
  | jq '.dashboard' > tfdrift-custom.json
```

### Import Dashboard

1. **Dashboards** → **Import**
2. **Upload JSON file** or **Paste JSON**
3. **Select folder**: TFDrift-Falco
4. **Import**

### Create Dashboard Snapshot

```
Dashboard → Share → Snapshot
- Set expiration: Never, 1 hour, 1 day, 1 week
- Publish to: snapshots.raintank.io (public) or local (private)
- Copy link
```

---

## Common Customizations

### 1. Add Team-specific Dashboard

Create a dashboard for your team's resources:

```logql
# DevOps team resources (by naming convention)
{job="tfdrift-falco"} | json | resource_id=~".*-devops-.*"

# Security team focus (IAM, SG, KMS)
{job="tfdrift-falco"} | json | resource_type=~"aws_iam_.*|aws_security_group|aws_kms_.*"

# Production environment only
{job="tfdrift-falco"} | json | resource_id=~".*-prod-.*"
```

### 2. Add SLA Tracking

Create panel showing drift response time:

```
Panel Title: "Drift Response SLA"
Threshold: 15 minutes
Query: Time since last critical drift
Visualization: Gauge (0-100%)
```

### 3. Add Compliance Dashboard

Track compliance-related resources:

```logql
# Encryption drifts
{job="tfdrift-falco"} | json | line_match_regex "encrypt|kms"

# Public access drifts
{job="tfdrift-falco"} | json | line_match_regex "public|PublicRead"

# Logging drifts
{job="tfdrift-falco"} | json | line_match_regex "logging|CloudTrail"
```

### 4. Add Change History Panel

Full audit trail with details:

```logql
{job="tfdrift-falco"} | json |
  line_format "{{.timestamp}} | {{.severity}} | {{.resource_type}} | {{.resource_id}} | {{.changed_by}} | {{.diff}}"
```

Visualization: **Logs panel** with:
- Time column
- Severity highlighting
- Expandable rows for diff details

### 5. Add Drift Velocity

Show rate of change over time:

```logql
# Drifts per hour
rate(count_over_time({job="tfdrift-falco"} | json [$__range])[1h])

# Acceleration (drifts per hour compared to previous hour)
deriv(count_over_time({job="tfdrift-falco"} | json [1h])[5m:1m])
```

### 6. Add Resource Health Score

Calculate health based on drift count:

```
Formula: 100 - (drift_count * 5)
Query: 100 - (count_over_time({job="tfdrift-falco"} | json | resource_id="$resource" [24h]) * 5)
Visualization: Gauge (0-100)
Thresholds:
  - 80-100: Green (healthy)
  - 50-80: Yellow (warning)
  - 0-50: Red (unhealthy)
```

### 7. Add Annotations

Mark important events on timelines:

```
Dashboard → Settings → Annotations → + Add annotation query

Name: Deployments
Data source: Loki
Query: {job="deployment-logs"} | json | status="success"
Tag keys: version, environment
```

Annotations appear as vertical lines on time series charts.

### 8. Add Links to External Tools

```
Panel → Panel options → Links → + Add link

Title: "View in AWS Console"
URL: https://console.aws.amazon.com/ec2/v2/home?region=us-east-1#Instances:instanceId=${__field.resource_id}

Title: "Terraform State"
URL: https://your-terraform-ui.com/resources/${__field.resource_id}
```

---

## Advanced Customizations

### Custom Loki Query Functions

Create reusable query templates:

```bash
# ~/.grafana/query-templates/tfdrift.txt

# Template: Critical in last N minutes
critical_last_n_min(n) =
  count_over_time({job="tfdrift-falco"} | json | severity="critical" [${n}m])

# Template: Top N resources
top_n_resources(n) =
  topk(${n}, sum by (resource_id) (count_over_time({job="tfdrift-falco"} | json [$__range])))
```

### Custom Panel Plugins

Install additional visualization plugins:

```bash
# Inside Grafana container
grafana-cli plugins install grafana-worldmap-panel
grafana-cli plugins install grafana-piechart-panel
grafana-cli plugins install grafana-clock-panel

# Restart Grafana
docker-compose restart grafana
```

### Dashboard Templating

Create dashboard templates for multiple environments:

```json
{
  "templating": {
    "list": [
      {
        "name": "environment",
        "type": "custom",
        "query": "prod,staging,dev",
        "current": {
          "value": "prod"
        }
      }
    ]
  }
}
```

Use in queries:
```logql
{job="tfdrift-falco", environment="$environment"} | json
```

---

## Troubleshooting

### Query Too Slow

```
Problem: Query takes >5 seconds
Solutions:
1. Reduce time range: Use $__range instead of [7d]
2. Add more specific filters: | resource_type="aws_s3_bucket"
3. Use label filters before JSON parsing: {severity="critical"} | json
4. Avoid regex when possible: Use = instead of =~
```

### Panel Not Updating

```
Problem: Panel shows old data
Solutions:
1. Check auto-refresh: Dashboard settings → Auto refresh → 30s
2. Check data source cache: Data source settings → Cache → Clear
3. Hard refresh: Shift + Cmd + R (Mac) or Ctrl + Shift + R (Windows)
4. Check time range: Ensure time range includes recent data
```

### Variables Not Working

```
Problem: Variable shows "No data"
Solutions:
1. Check data source connection
2. Verify query syntax: label_values(severity) not label_values({job="..."}, severity)
3. Check variable dependencies: Ensure chained variables load in order
4. Test query in Explore view first
```

---

## Best Practices

1. **Start with Templates**: Copy existing panels and modify
2. **Test Queries**: Use Explore view before adding to dashboards
3. **Use Variables**: Make dashboards flexible with variables
4. **Document Changes**: Add panel descriptions explaining custom logic
5. **Version Control**: Export JSON and commit to git
6. **Monitor Performance**: Keep queries under 2-3 seconds
7. **Organize Folders**: Group related dashboards
8. **Use Consistent Naming**: Follow naming conventions for variables
9. **Add Helpful Links**: Link to runbooks and documentation
10. **Regular Backups**: Export dashboards weekly

---

## Resources

- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
- [Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)
- [Panel Plugin Library](https://grafana.com/grafana/plugins/)

---

**Need Help?**

- Create an issue: https://github.com/your-org/tfdrift-falco/issues
- Slack: #tfdrift-falco
- Email: support@your-org.com
