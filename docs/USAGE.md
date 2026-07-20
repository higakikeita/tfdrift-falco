# TFDrift-Falco Usage Guide

## Quick Start

### 1. Basic Drift Detection (Display Only)

This mode detects unmanaged resources and displays terraform import commands, but does not execute them.

**Config:** `examples/config-with-autoimport.yaml`
```yaml
auto_import:
  enabled: false  # Display only, no execution
```

**Run:**
```bash
tfdrift --config examples/config-with-autoimport.yaml
```

**Output when unmanaged resource detected:**
```
⚠️  UNMANAGED RESOURCE DETECTED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Resource:
   Type: aws_iam_role
   ID:   production-api-role

💡 Recommendation:
   terraform import aws_iam_role.production_api_role production-api-role
```

---

## Auto-Import Modes

### Mode 1: Manual Approval (Recommended for Production)

Enable auto-import with manual approval prompts.

**Config:**
```yaml
auto_import:
  enabled: true
  require_approval: true  # Prompt for approval
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"
  allowed_resources:
    - "aws_iam_role"
    - "aws_iam_policy"
```

**Run:**
```bash
tfdrift --config config.yaml --interactive
```

**Interactive Flow:**
```
🔔 IMPORT APPROVAL REQUIRED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Resource Type: aws_iam_role
🆔 Resource ID:   production-api-role
📝 Resource Name: production_api_role (auto-generated)
👤 Detected By:   john.doe@company.com (arn:aws:iam::123456789012:user/john.doe)
🕐 Detected At:   2025-01-15T14:23:45Z

🔄 Changes:
   role_name: production-api-role
   assume_role_policy: {...}

💻 Import Command:
   terraform import aws_iam_role.production_api_role production-api-role

❓ Approve this import? [y/N]: y
✅ Import approved!
🚀 Executing: terraform import aws_iam_role.production_api_role production-api-role
✅ Import successful!

📄 Generated Terraform code:
resource "aws_iam_role" "production_api_role" {
  name = "production-api-role"
  # ... attributes ...
}

💡 Save this to: ./infrastructure/imported/aws_iam_role_production_api_role.tf
```

---

### Mode 2: Auto-Approval with Whitelist (Development/Staging)

Automatically import resources from a whitelist without prompting.

**Config:**
```yaml
auto_import:
  enabled: true
  require_approval: false  # Auto-approve
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"
  allowed_resources:
    - "aws_iam_role"
    - "aws_iam_policy"
    # Only these types will be auto-imported
```

**Run:**
```bash
tfdrift --config config.yaml
```

**Behavior:**
- Resources in `allowed_resources` → Auto-imported immediately
- Resources NOT in list → Skipped (display only)

---

### Mode 3: Full Auto (Testing Only - Not Recommended)

Automatically import ALL unmanaged resources without approval.

**Config:**
```yaml
auto_import:
  enabled: true
  require_approval: false
  allowed_resources: []  # Empty = all resources
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"
```

**⚠️ WARNING:** Use this ONLY in isolated test environments!

---

## CLI Commands

### Drift Detection

```bash
# Basic run with config
tfdrift --config config.yaml

# Interactive mode (for manual approval)
tfdrift --config config.yaml --interactive

# Dry-run mode (no actual imports or notifications)
tfdrift --config config.yaml --dry-run

# Daemon mode (background process)
tfdrift --config config.yaml --daemon
```

### Approval Management

```bash
# List pending approval requests
tfdrift approval list

# Approve a specific request
tfdrift approval approve <request-id>

# Reject a request with reason
tfdrift approval reject <request-id> --reason "Not needed"

# Clean up expired requests
tfdrift approval cleanup --older-than 24h
```

**Note:** Approval commands currently require a running TFDrift-Falco instance. For now, use `--interactive` mode instead.

---

## Configuration Options

### Complete Config Structure

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state"
      s3_key: "terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

auto_import:
  enabled: true
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"
  allowed_resources:
    - "aws_iam_role"
    - "aws_iam_policy"
  require_approval: true

drift_rules:
  - name: "IAM Role Trust Policy Change"
    resource_types:
      - "aws_iam_role"
    watched_attributes:
      - "assume_role_policy"
    severity: "critical"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#infra-alerts"

logging:
  level: "info"
  format: "json"
```

---

## Best Practices

### 🟢 Recommended

1. **Use manual approval in production**
   ```yaml
   require_approval: true
   ```

2. **Use whitelist for auto-approval**
   ```yaml
   allowed_resources:
     - "aws_iam_role"
     - "aws_iam_policy"
   ```

3. **Start with dry-run**
   ```bash
   tfdrift --config config.yaml --dry-run
   ```

4. **Review generated .tf files**
   - Auto-generated code is basic
   - Add tags, descriptions, and complex attributes manually

### 🔴 Not Recommended

1. **Full auto-approval in production**
   ```yaml
   # DON'T DO THIS IN PRODUCTION
   require_approval: false
   allowed_resources: []
   ```

2. **Auto-importing EC2 instances**
   ```yaml
   # DON'T - state file will be huge
   allowed_resources:
     - "aws_instance"
   ```

---

## Troubleshooting

### Import fails with "terraform not initialized"

**Solution:**
```bash
cd ./infrastructure
terraform init
```

### Generated code is incomplete

**Expected:** Auto-generated code only includes basic attributes.

**Action:** Manually add:
- Complex nested blocks
- Tags
- Dependencies
- Custom attributes

### Approval prompt not appearing

**Solution:** Use `--interactive` flag:
```bash
tfdrift --config config.yaml --interactive
```

### Resource not being auto-imported

**Check:**
1. Is `auto_import.enabled: true`?
2. Is resource type in `allowed_resources` (if set)?
3. Is approval workflow properly configured?

**Debug:**
```bash
tfdrift --config config.yaml --dry-run
# Check log output for approval workflow messages
```

---

## Examples

See the `examples/` directory for:
- `config-with-autoimport.yaml` - Full configuration example

> Visualization is provided by the built-in React web UI (`ui/`, served by the API server). The earlier Grafana/Loki dashboards were removed in favor of the purpose-built UI — see `ui/docs/qiita-article.md` for the rationale.

For detailed auto-import documentation, see `docs/auto-import-guide.md`.
