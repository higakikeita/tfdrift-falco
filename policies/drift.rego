# METADATA
# title: TFDrift-Falco Core Drift Policy
# description: Classifies detected drifts into allow/alert/remediate/deny.
# authors:
#   - TFDrift-Falco

package tfdrift

import rego.v1

# Default decision: alert on any drift
default decision := "alert"

default reason := "drift detected, no specific policy matched"

# ---------- ALLOW rules (suppress false positives) ----------

# Auto Scaling group changes are expected
decision := "allow" if {
	input.resource_type == "aws_autoscaling_group"
	input.attribute in {"desired_capacity", "instances"}
}

reason := "Auto Scaling changes are expected operational behaviour" if {
	input.resource_type == "aws_autoscaling_group"
	input.attribute in {"desired_capacity", "instances"}
}

# ECS task count changes from auto-scaling
decision := "allow" if {
	input.resource_type == "aws_ecs_service"
	input.attribute == "desired_count"
}

reason := "ECS desired_count changes from auto-scaling are expected" if {
	input.resource_type == "aws_ecs_service"
	input.attribute == "desired_count"
}

# Tag-only changes are low priority
decision := "allow" if {
	input.attribute == "tags"
	input.severity != "critical"
}

reason := "Tag-only changes are informational" if {
	input.attribute == "tags"
	input.severity != "critical"
}

# ---------- REMEDIATE rules (auto-fix) ----------

# Security group ingress opened to 0.0.0.0/0 should be auto-remediated
decision := "remediate" if {
	input.resource_type == "aws_security_group"
	input.attribute == "ingress"
	_contains_open_cidr(input.new_value)
}

reason := "Security group opened to 0.0.0.0/0 — auto-remediate" if {
	input.resource_type == "aws_security_group"
	input.attribute == "ingress"
	_contains_open_cidr(input.new_value)
}

severity := "critical" if {
	input.resource_type == "aws_security_group"
	input.attribute == "ingress"
	_contains_open_cidr(input.new_value)
}

# S3 bucket public access enabled → remediate
decision := "remediate" if {
	input.resource_type == "aws_s3_bucket"
	input.attribute in {"acl", "public_access_block"}
	_is_public_access(input.new_value)
}

reason := "S3 bucket public access detected — auto-remediate" if {
	input.resource_type == "aws_s3_bucket"
	input.attribute in {"acl", "public_access_block"}
	_is_public_access(input.new_value)
}

# ---------- DENY rules (policy violations that must be escalated) ----------

# IAM policy changes by unknown users
decision := "deny" if {
	startswith(input.resource_type, "aws_iam")
	input.user_identity.user_name == ""
}

reason := "IAM changes by unidentified user — policy violation" if {
	startswith(input.resource_type, "aws_iam")
	input.user_identity.user_name == ""
}

severity := "critical" if {
	startswith(input.resource_type, "aws_iam")
	input.user_identity.user_name == ""
}

# Encryption disabled on any resource
decision := "deny" if {
	input.attribute in {"encrypted", "kms_key_id", "server_side_encryption"}
	_encryption_disabled(input.new_value)
}

reason := "Encryption was disabled — policy violation" if {
	input.attribute in {"encrypted", "kms_key_id", "server_side_encryption"}
	_encryption_disabled(input.new_value)
}

# ---------- Unmanaged resources ----------

# Unmanaged resources in production should be denied
decision := "deny" if {
	input.type == "unmanaged"
	startswith(input.resource_type, "aws_iam")
}

reason := "Unmanaged IAM resource created outside Terraform — policy violation" if {
	input.type == "unmanaged"
	startswith(input.resource_type, "aws_iam")
}

# ---------- Helper rules ----------

_contains_open_cidr(val) if {
	is_string(val)
	contains(val, "0.0.0.0/0")
}

_contains_open_cidr(val) if {
	is_array(val)
	some item in val
	is_string(item)
	contains(item, "0.0.0.0/0")
}

_is_public_access(val) if {
	val == "public-read"
}

_is_public_access(val) if {
	val == "public-read-write"
}

_encryption_disabled(val) if {
	val == false
}

_encryption_disabled(val) if {
	val == ""
}

_encryption_disabled(val) if {
	is_null(val)
}

# ---------- Labels for routing ----------

labels := {"team": "security"} if {
	decision == "deny"
}

labels := {"team": "platform"} if {
	decision == "remediate"
}

# ---------- Suppressors ----------

suppressors contains "autoscaling" if {
	input.resource_type == "aws_autoscaling_group"
}

suppressors contains "ecs-scaling" if {
	input.resource_type == "aws_ecs_service"
	input.attribute == "desired_count"
}
