# TFDrift-Falco Alert Configuration Guide

This guide explains how to set up alerts for TFDrift-Falco drift detection events in Grafana.

## Alert Overview

We provide 6 pre-configured alert rules to monitor drift events:

| Alert Name | Severity | Trigger Condition | Use Case |
|------------|----------|-------------------|----------|
| **Critical Drift Detected** | Critical | >1 critical event in 5 min | Immediate response required |
| **High Severity Drift** | High | >3 high events in 10 min | Multiple high-impact drifts |
| **Security Group Drift** | High | Any SG drift in 5 min | Network security changes |
| **IAM Policy/Role Drift** | Critical | Any IAM drift in 5 min | Identity & access changes |
| **S3 Public Access Drift** | Critical | S3 public access change | Data exposure risk |
| **Excessive Drift Rate** | Medium | >10 events in 1 hour | Systemic issues |

## Quick Setup (UI-Based)

### Option 1: Import Alert Rules via UI

1. **Access Grafana**
   ```bash
   open http://localhost:3000
   ```
   Login: `admin` / `admin`

2. **Navigate to Alerting**
   - Click **Alerting** (bell icon) in left sidebar
   - Click **Alert rules**
   - Click **+ New alert rule**

3. **Create Alert: Critical Drift Detected**

   **Section A: Enter alert rule name**
   - Name: `Critical Drift Detected`
   - Folder: `TFDrift-Falco` (create new if needed)

   **Section B: Set a query and alert condition**
   - Data source: `Loki`
   - Query A:
     ```logql
     count_over_time({job="tfdrift-falco"} | json | severity="critical" [5m])
     ```
   - Add Expression: **Reduce** (B)
     - Function: `Last`
     - Input: A
   - Add Expression: **Threshold** (C)
     - Input: B
     - IS ABOVE: `1`
   - Set as alert condition: C

   **Section C: Set evaluation behavior**
   - Folder: `TFDrift-Falco`
   - Evaluation group: `tfdrift-alerts` (create new)
   - Evaluation interval: `1m`
   - For: `1m`

   **Section D: Add annotations**
   - Description: `Critical severity drift detected in the last 5 minutes`
   - Summary: `Critical drift alert`

   **Section E: Add labels**
   - `severity`: `critical`
   - `team`: `security`

   Click **Save rule and exit**

4. **Repeat for other alerts** using the queries below

### Alert Queries Reference

#### Alert 2: High Severity Drift
```logql
# Query A
count_over_time({job="tfdrift-falco"} | json | severity="high" [10m])
# Threshold: >3
# For: 2m
```

#### Alert 3: Security Group Drift
```logql
# Query A
count_over_time({job="tfdrift-falco"} | json | resource_type="aws_security_group" [5m])
# Threshold: >1
# For: 1m
# Labels: severity=high, resource_type=aws_security_group
```

#### Alert 4: IAM Policy/Role Drift
```logql
# Query A
count_over_time({job="tfdrift-falco"} | json | resource_type=~"aws_iam_.*" [5m])
# Threshold: >1
# For: 1m
# Labels: severity=critical, resource_type=aws_iam
```

#### Alert 5: S3 Public Access Drift
```logql
# Query A
count_over_time({job="tfdrift-falco"} | json | resource_type="aws_s3_bucket" | line_match_regex "public_access_block|block_public" [5m])
# Threshold: >1
# For: 30s
# Labels: severity=critical, resource_type=aws_s3_bucket
```

#### Alert 6: Excessive Drift Rate
```logql
# Query A
count_over_time({job="tfdrift-falco"} | json | action="drift_detected" [1h])
# Threshold: >10
# For: 5m
# Labels: severity=medium
```

## Notification Setup

### 1. Slack Notifications

1. **Create Slack Webhook**
   - Go to https://api.slack.com/apps
   - Create new app → Incoming Webhooks
   - Copy webhook URL (e.g., `https://hooks.slack.com/services/...`)

2. **Configure in Grafana**
   - Alerting → Contact points
   - **+ Add contact point**
   - Name: `slack-tfdrift`
   - Integration: `Slack`
   - Webhook URL: Paste your webhook URL
   - Optional: Customize message template
   - **Test** → **Save contact point**

3. **Create Notification Policy**
   - Alerting → Notification policies
   - **+ New specific policy**
   - Matching labels:
     - `severity` = `critical`
   - Contact point: `slack-tfdrift`
   - **Save policy**

### 2. Email Notifications

1. **Configure SMTP in docker-compose.yaml**
   ```yaml
   environment:
     - GF_SMTP_ENABLED=true
     - GF_SMTP_HOST=smtp.gmail.com:587
     - GF_SMTP_USER=your-email@gmail.com
     - GF_SMTP_PASSWORD=your-app-password
     - GF_SMTP_FROM_ADDRESS=grafana@example.com
     - GF_SMTP_FROM_NAME=TFDrift Grafana
   ```

2. **Restart Grafana**
   ```bash
   docker-compose restart grafana
   ```

3. **Add Email Contact Point**
   - Alerting → Contact points
   - **+ Add contact point**
   - Name: `email-tfdrift`
   - Integration: `Email`
   - Addresses: `security@example.com,ops@example.com`
   - **Test** → **Save contact point**

### 3. Webhook Notifications

For integration with PagerDuty, Opsgenie, or custom systems:

1. **Add Webhook Contact Point**
   - Alerting → Contact points
   - **+ Add contact point**
   - Name: `webhook-tfdrift`
   - Integration: `Webhook`
   - URL: Your webhook endpoint
   - HTTP Method: `POST`
   - **Test** → **Save contact point**

## Notification Policy Examples

### Example 1: Critical to Slack, High to Email

```
Default policy → default-contact-point

├─ Specific policy 1 (continue)
│  Matching labels: severity = critical
│  Contact point: slack-tfdrift
│  Group by: alertname
│  Group wait: 10s
│  Repeat interval: 30m
│
└─ Specific policy 2
   Matching labels: severity = high
   Contact point: email-tfdrift
   Group by: alertname, resource_type
   Group wait: 30s
   Repeat interval: 2h
```

### Example 2: Security Team Gets All Security Resources

```
Default policy → default-contact-point

└─ Specific policy
   Matching labels: resource_type =~ aws_security_group|aws_iam_.*
   Contact point: slack-security-team
   Group by: resource_type
   Group wait: 15s
   Repeat interval: 1h
```

## Testing Alerts

### Method 1: Generate Test Events

```bash
# Add test drift events to trigger alerts
cat >> /path/to/tfdrift-falco/dashboards/grafana/sample-logs/current-drift-events.jsonl << 'EOF'
{"timestamp":"$(date -u +%Y-%m-%dT%H:%M:%SZ)","resource_type":"aws_security_group","resource_id":"sg-test-alert","changed_by":"test-user","severity":"critical","diff":{"ingress":{"expected":["443/tcp"],"actual":["443/tcp","22/tcp"]}},"action":"drift_detected"}
EOF

# Wait 1-2 minutes for alert to fire
```

### Method 2: Use Grafana Test Button

1. Go to **Alerting** → **Alert rules**
2. Find your alert rule
3. Click **...** (3 dots) → **Test**
4. Verify alert fires correctly

### Method 3: Lower Thresholds Temporarily

Change threshold from `>1` to `>0` to trigger on any event.

## Alert Monitoring

### View Active Alerts

- **Alerting** → **Alert rules**: See all configured rules
- **Alerting** → **Alert list**: See currently firing alerts
- **Alerting** → **Silences**: Temporarily suppress alerts

### Check Alert Status

```bash
# Get all alert rules
curl -u admin:admin http://localhost:3000/api/ruler/grafana/api/v1/rules | jq

# Get alert status
curl -u admin:admin http://localhost:3000/api/v1/rules | jq
```

## Troubleshooting

### Alerts Not Firing

1. **Check query returns data**
   ```bash
   # In Grafana Explore, run:
   count_over_time({job="tfdrift-falco"} | json | severity="critical" [5m])
   ```

2. **Verify evaluation interval**
   - Must be ≥ 1m for Loki queries
   - Check Alerting → Alert rules → Evaluation group

3. **Check alert state**
   - Go to alert rule
   - View **State history** tab
   - Look for errors

### Notifications Not Sending

1. **Test contact point**
   - Alerting → Contact points
   - Click **Test** button
   - Check for error messages

2. **Check notification policy**
   - Alerting → Notification policies
   - Verify labels match alert labels
   - Check contact point is selected

3. **View notification history**
   - Alerting → Contact points
   - Click contact point name
   - View **Notification history**

### SMTP Email Issues

```bash
# Check Grafana logs
docker-compose logs grafana | grep -i smtp

# Common issues:
# - Wrong SMTP host/port
# - Authentication failure (use app password for Gmail)
# - Firewall blocking port 587/465
```

## Best Practices

1. **Start with UI setup first** - Easier to test and iterate
2. **Test each alert individually** before enabling all
3. **Use appropriate thresholds** - Avoid alert fatigue
4. **Group related alerts** - Use notification policies effectively
5. **Set appropriate repeat intervals** - Balance responsiveness vs noise
6. **Monitor alert volume** - Adjust thresholds if too noisy
7. **Document your policies** - Help team understand alert routing

## Environment Variables for docker-compose

Create a `.env` file in the `dashboards/grafana/` directory:

```bash
# Slack webhook
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Email addresses (comma-separated)
ALERT_EMAIL_ADDRESSES=security@example.com,ops@example.com

# Generic webhook
WEBHOOK_URL=https://your-webhook-endpoint.com/alerts

# SMTP configuration
GF_SMTP_ENABLED=true
GF_SMTP_HOST=smtp.gmail.com:587
GF_SMTP_USER=your-email@gmail.com
GF_SMTP_PASSWORD=your-app-password
GF_SMTP_FROM_ADDRESS=grafana@example.com
GF_SMTP_FROM_NAME=TFDrift Grafana
```

Then restart:
```bash
docker-compose down
docker-compose up -d
```

## Next Steps

1. Set up your first alert using the UI
2. Configure at least one notification channel (Slack or email)
3. Generate test events to verify alerts fire
4. Create notification policies for your team structure
5. Document your alert thresholds and response procedures

For more information, see:
- [Grafana Alerting Documentation](https://grafana.com/docs/grafana/latest/alerting/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
