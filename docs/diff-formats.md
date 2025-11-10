# Diff Format Examples

TFDrift-Falco provides multiple diff formats to display the differences between Terraform state and actual runtime configuration.

## Table of Contents

1. [Console Format (Colored)](#console-format-colored)
2. [Unified Diff Format](#unified-diff-format)
3. [Side-by-Side Format](#side-by-side-format)
4. [Markdown Format](#markdown-format)
5. [JSON Format](#json-format)

---

## Console Format (Colored)

The console format provides a human-readable, colorized output perfect for terminal viewing.

### Example Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚨 DRIFT DETECTED: aws_instance.webserver
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Severity: CRITICAL

📦 Resource:
  Type:       aws_instance
  Name:       webserver
  ID:         i-0abcd1234efgh5678

🔄 Changed Attribute:
  disable_api_termination

📝 Value Change:
  - true  →  + false

👤 Changed By:
  User:       admin-user@example.com
  Type:       IAMUser
  ARN:        arn:aws:iam::123456789012:user/admin-user
  Account:    123456789012

⏰ Timestamp:
  2025-01-15T10:35:10Z

📋 Matched Rules:
  • EC2 Instance Termination Protection

📄 Terraform Code:
  # Current Terraform Definition:
  resource "aws_instance" "webserver" {
    disable_api_termination = true
    # ... other attributes ...
  }

  # Actual Runtime Configuration:
  resource "aws_instance" "webserver" {
    disable_api_termination = false
    # ... other attributes ...
  }

💡 Recommendations:
  1. Review the change with the user who made it
  2. Determine if the change is authorized
  3. Update Terraform code if the change is intentional:
     terraform plan && terraform apply
  4. Or revert the manual change to match IaC:
     terraform apply -target=aws_instance.webserver

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Complex Value Changes

For complex attributes like IAM policies or security group rules:

```
📝 Value Change:
  - Old Value:
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "s3:GetObject",
          "Resource": "arn:aws:s3:::my-bucket/*"
        }
      ]
    }

  + New Value:
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "*",
          "Resource": "*"
        }
      ]
    }
```

---

## Unified Diff Format

Git-style unified diff format, familiar to developers.

### Example Output

```diff
--- terraform/aws_instance.webserver	(Terraform State)
+++ runtime/aws_instance.webserver	(Actual Configuration)
@@ -1,1 +1,1 @@
-disable_api_termination = true
+disable_api_termination = false
```

### Complex IAM Policy Diff

```diff
--- terraform/aws_iam_policy.s3_access	(Terraform State)
+++ runtime/aws_iam_policy.s3_access	(Actual Configuration)
@@ -1,9 +1,9 @@
 {
   "Version": "2012-10-17",
   "Statement": [
     {
       "Effect": "Allow",
-      "Action": "s3:GetObject",
-      "Resource": "arn:aws:s3:::my-bucket/*"
+      "Action": "*",
+      "Resource": "*"
     }
   ]
 }
```

---

## Side-by-Side Format

Two-column comparison for easy visual comparison.

### Example Output

```
Terraform State                          | Actual Configuration
─────────────────────────────────────────┼─────────────────────────────────────────
  disable_api_termination = true         │   disable_api_termination = false
  instance_type = "t2.micro"             │   instance_type = "t2.micro"
  monitoring = false                     │   monitoring = false
```

### Complex Security Group Rules

```
Terraform State                          | Actual Configuration
─────────────────────────────────────────┼─────────────────────────────────────────
  ingress {                              │   ingress {
    from_port   = 443                    │     from_port   = 443
    to_port     = 443                    │     to_port     = 443
    protocol    = "tcp"                  │     protocol    = "tcp"
-   cidr_blocks = ["10.0.0.0/8"]         │ +   cidr_blocks = ["0.0.0.0/0"]
  }                                      │   }
```

---

## Markdown Format

Perfect for Slack, GitHub Issues, or documentation.

### Example Output

````markdown
## 🚨 Drift Detected: `aws_instance.webserver`

**Severity:** 🔴 **CRITICAL**

**Changed Attribute:** `disable_api_termination`

### Value Change

```diff
- true
+ false
```

### Changed By

- **User:** admin-user@example.com
- **Account:** 123456789012
- **Time:** 2025-01-15T10:35:10Z

### Terraform State

```hcl
resource "aws_instance" "webserver" {
  disable_api_termination = true
  # ... other attributes ...
}
```

### Actual Configuration

```hcl
resource "aws_instance" "webserver" {
  disable_api_termination = false
  # ... other attributes ...
}
```

### Recommended Actions

- [ ] Review change with user
- [ ] Update Terraform code if intentional
- [ ] Run `terraform apply -target=aws_instance.webserver` to revert
````

### Slack Rendering

When sent to Slack, this renders as:

![Slack Example](../examples/images/slack-drift-example.png)

---

## JSON Format

Machine-readable format for API integrations and SIEM systems.

### Example Output

```json
{
  "severity": "critical",
  "resource_type": "aws_instance",
  "resource_name": "webserver",
  "resource_id": "i-0abcd1234efgh5678",
  "attribute": "disable_api_termination",
  "change": {
    "old_value": true,
    "new_value": false
  },
  "user": {
    "name": "admin-user@example.com",
    "type": "IAMUser",
    "arn": "arn:aws:iam::123456789012:user/admin-user",
    "account_id": "123456789012",
    "principal_id": "AIDAI23456EXAMPLE"
  },
  "timestamp": "2025-01-15T10:35:10Z",
  "matched_rules": [
    "EC2 Instance Termination Protection"
  ],
  "terraform_code": {
    "state_definition": "resource \"aws_instance\" \"webserver\" {\n  disable_api_termination = true\n  # ... other attributes ...\n}",
    "actual_config": "resource \"aws_instance\" \"webserver\" {\n  disable_api_termination = false\n  # ... other attributes ...\n}"
  }
}
```

### Usage in API

```bash
# Get JSON format
curl -X GET http://tfdrift-api:8080/drifts/latest \
  -H "Accept: application/json" | jq .

# Send to SIEM
curl -X POST https://siem.example.com/api/events \
  -H "Content-Type: application/json" \
  -d @drift-event.json
```

---

## Real-World Examples

### Example 1: S3 Bucket Policy Change

**Scenario:** Someone manually changed an S3 bucket policy to allow public access.

**Console Output:**
```
🚨 DRIFT DETECTED: aws_s3_bucket_policy.data_bucket
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Severity: CRITICAL

📝 Value Change:
  - Old Policy:
    {
      "Principal": {"AWS": "arn:aws:iam::123456789012:role/DataAccess"},
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::data-bucket/*"
    }

  + New Policy:
    {
      "Principal": "*",
      "Action": "s3:*",
      "Resource": "arn:aws:s3:::data-bucket/*"
    }

⚠️  WARNING: Bucket now allows public access!
```

### Example 2: Lambda Function Memory Increase

**Scenario:** Developer increased Lambda memory allocation to debug performance issue.

**Console Output:**
```
🚨 DRIFT DETECTED: aws_lambda_function.api_handler
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Severity: MEDIUM

📝 Value Change:
  memory_size: 512 → 3008

💰 Cost Impact: ~$25/month increase

💡 Recommendations:
  - Review if increased memory is necessary
  - Update Terraform if performance improvement confirmed
  - Consider reverting if temporary debug change
```

### Example 3: Security Group Rule Addition

**Scenario:** Network admin added temporary SSH access rule.

**Side-by-Side Output:**
```
Terraform State                          | Actual Configuration
─────────────────────────────────────────┼─────────────────────────────────────────
  ingress {                              │   ingress {
    from_port = 443                      │     from_port = 443
    to_port = 443                        │     to_port = 443
    protocol = "tcp"                     │     protocol = "tcp"
    cidr_blocks = ["10.0.0.0/8"]         │     cidr_blocks = ["10.0.0.0/8"]
  }                                      │   }
                                         │ + ingress {
                                         │ +   from_port = 22
                                         │ +   to_port = 22
                                         │ +   protocol = "tcp"
                                         │ +   cidr_blocks = ["0.0.0.0/0"]
                                         │ + }
```

---

## Configuration

### Enable/Disable Formats

In `config.yaml`:

```yaml
output:
  formats:
    console:
      enabled: true
      colors: true

    unified_diff:
      enabled: true

    side_by_side:
      enabled: false

    markdown:
      enabled: true

    json:
      enabled: true
      file: "/var/log/tfdrift/drifts.json"
```

### Color Customization

```yaml
output:
  colors:
    critical: "red"
    high: "yellow"
    medium: "blue"
    low: "green"
    added: "green"
    removed: "red"
    changed: "yellow"
```

---

## Programmatic Access

### Go API

```go
import "github.com/keitahigaki/tfdrift-falco/pkg/diff"

// Create formatter
formatter := diff.NewFormatter(true) // Enable colors

// Format in different ways
consoleDiff := formatter.FormatConsole(alert)
unifiedDiff := formatter.FormatUnifiedDiff(alert)
sideBySide := formatter.FormatSideBySide(alert)
markdown := formatter.FormatMarkdown(alert)
jsonStr, _ := formatter.FormatJSON(alert)

fmt.Println(consoleDiff)
```

### CLI Usage

```bash
# Default console format
tfdrift --config config.yaml

# JSON output
tfdrift --config config.yaml --format json

# Unified diff format
tfdrift --config config.yaml --format diff

# Markdown format (for piping to Slack)
tfdrift --config config.yaml --format markdown | slack-cli post
```

---

## Best Practices

### 1. Console Output for Interactive Use

Use colorized console format when running interactively:
```bash
tfdrift --config config.yaml
```

### 2. JSON for Automation

Use JSON format when integrating with other tools:
```bash
tfdrift --format json | jq '.severity == "critical"' | while read drift; do
  # Send to PagerDuty
  curl -X POST https://api.pagerduty.com/incidents ...
done
```

### 3. Markdown for Notifications

Use Markdown format for Slack/Discord/GitHub:
```yaml
notifications:
  slack:
    format: "markdown"
```

### 4. Unified Diff for Git Integration

Use unified diff format for committing to audit repository:
```bash
tfdrift --format diff > drifts/$(date +%Y-%m-%d-%H%M%S).diff
git add drifts/*.diff
git commit -m "Detected drifts: $(date)"
```

---

## Future Enhancements

### Planned Features

- [ ] HTML format with interactive visualization
- [ ] CSV format for Excel/Google Sheets
- [ ] Annotated screenshots (for UI changes)
- [ ] Video recording of change timeline
- [ ] AI-generated natural language explanations

---

**Document Version:** 1.0
**Last Updated:** 2025-01-15
**Maintainer:** Keita Higaki
