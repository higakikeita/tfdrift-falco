# METADATA
# title: GCP-specific drift policies
# description: Policy rules for Google Cloud Platform resources.

package tfdrift

import rego.v1

# GCP managed instance group size changes from autoscaler
decision := "allow" if {
	input.resource_type == "google_compute_instance_group_manager"
	input.attribute == "target_size"
}

reason := "GCP managed instance group autoscaler change" if {
	input.resource_type == "google_compute_instance_group_manager"
	input.attribute == "target_size"
}

# GCP firewall rule opened to 0.0.0.0/0
decision := "remediate" if {
	input.resource_type == "google_compute_firewall"
	input.attribute == "source_ranges"
	_contains_open_cidr(input.new_value)
}

reason := "GCP firewall opened to 0.0.0.0/0 — auto-remediate" if {
	input.resource_type == "google_compute_firewall"
	input.attribute == "source_ranges"
	_contains_open_cidr(input.new_value)
}

severity := "critical" if {
	input.resource_type == "google_compute_firewall"
	input.attribute == "source_ranges"
	_contains_open_cidr(input.new_value)
}

# GCP IAM binding changes by service accounts
decision := "deny" if {
	input.resource_type == "google_project_iam_binding"
	input.user_identity.user_name == ""
}

reason := "GCP IAM binding changed by unidentified user — policy violation" if {
	input.resource_type == "google_project_iam_binding"
	input.user_identity.user_name == ""
}
