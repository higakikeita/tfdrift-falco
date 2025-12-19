# AWS WAF (Web Application Firewall)

# WAF Web ACL
resource "aws_wafv2_web_acl" "main" {
  name  = "deepdrift-test-waf"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  # Rule 1: Rate limiting
  rule {
    name     = "rate-limit-rule"
    priority = 1

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitRule"
      sampled_requests_enabled   = true
    }
  }

  # Rule 2: AWS Managed Rules - Core Rule Set
  rule {
    name     = "aws-managed-core-rule-set"
    priority = 2

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedCoreRuleSet"
      sampled_requests_enabled   = true
    }
  }

  # Rule 3: AWS Managed Rules - Known Bad Inputs
  rule {
    name     = "aws-managed-known-bad-inputs"
    priority = 3

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesKnownBadInputsRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedKnownBadInputs"
      sampled_requests_enabled   = true
    }
  }

  # Rule 4: AWS Managed Rules - SQL Injection
  rule {
    name     = "aws-managed-sql-injection"
    priority = 4

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesSQLiRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedSQLInjection"
      sampled_requests_enabled   = true
    }
  }

  # Rule 5: Geographic blocking (example: block traffic from specific countries)
  rule {
    name     = "geo-blocking-rule"
    priority = 5

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = ["KP", "IR"] # North Korea, Iran
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "GeoBlockingRule"
      sampled_requests_enabled   = true
    }
  }

  # Rule 6: Custom IP Set blocking
  rule {
    name     = "ip-blacklist-rule"
    priority = 6

    action {
      block {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.blacklist.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "IPBlacklistRule"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "DeepDriftTestWAF"
    sampled_requests_enabled   = true
  }

  tags = {
    Name = "deepdrift-test-waf"
  }
}

# IP Set for blacklist
resource "aws_wafv2_ip_set" "blacklist" {
  name               = "deepdrift-test-ip-blacklist"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"

  # Example IPs (replace with actual IPs to block)
  addresses = [
    "192.0.2.0/24",   # Example IP range
    "198.51.100.0/24", # Example IP range
  ]

  tags = {
    Name = "deepdrift-test-ip-blacklist"
  }
}

# Associate WAF with ALB
resource "aws_wafv2_web_acl_association" "alb" {
  resource_arn = aws_lb.main.arn
  web_acl_arn  = aws_wafv2_web_acl.main.arn
}

# WAF Logging Configuration
resource "aws_wafv2_web_acl_logging_configuration" "main" {
  resource_arn            = aws_wafv2_web_acl.main.arn
  log_destination_configs = ["${aws_cloudwatch_log_group.waf.arn}:*"]

  redacted_fields {
    single_header {
      name = "authorization"
    }
  }

  redacted_fields {
    single_header {
      name = "cookie"
    }
  }
}

# CloudWatch Log Group for WAF
resource "aws_cloudwatch_log_group" "waf" {
  name              = "/aws/wafv2/deepdrift-test"
  retention_in_days = 7

  tags = {
    Name = "deepdrift-test-waf-logs"
  }
}

# CloudWatch Metric Alarm for WAF blocked requests
resource "aws_cloudwatch_metric_alarm" "waf_blocked_requests" {
  alarm_name          = "deepdrift-test-waf-blocked-requests"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "BlockedRequests"
  namespace           = "AWS/WAFV2"
  period              = "300"
  statistic           = "Sum"
  threshold           = "100"
  alarm_description   = "Alert when WAF blocks more than 100 requests in 5 minutes"
  treat_missing_data  = "notBreaching"

  dimensions = {
    WebACL = aws_wafv2_web_acl.main.name
    Region = "us-east-1"
    Rule   = "ALL"
  }

  tags = {
    Name = "deepdrift-test-waf-blocked-alarm"
  }
}
