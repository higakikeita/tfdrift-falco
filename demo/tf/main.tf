# TFDrift-Falco — demo infrastructure (OSS Summit Korea 2026)
#
# Deliberately tiny: one security group, one ingress rule, one private range.
# The point is that the audience can read the whole thing in ten seconds and
# then watch it get violated in the AWS console.
#
# ── Why inline `ingress` blocks and not aws_vpc_security_group_ingress_rule ──
# This matters for the demo to work at all. With inline blocks, Terraform owns
# the *entire* rule set of the security group, so a rule added by hand in the
# console shows up as a MODIFIED aws_security_group — which is the story we are
# telling. If we used separate aws_vpc_security_group_ingress_rule resources, a
# console-added rule would be a resource Terraform has never seen, and it would
# be reported as UNMANAGED instead. Both are real drift, but only the first one
# demonstrates "the resource I manage no longer matches what I wrote."

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-northeast-1"
}

# Use an existing VPC (this shared account has no default VPC and is at its VPC
# limit, so we don't create one). The SG is the only thing Terraform manages, so
# `terraform destroy` removes exactly the demo SG and nothing else.
data "aws_vpc" "demo" {
  tags = {
    Name = "tfdrift-lab-vpc"
  }
}

resource "aws_security_group" "web" {
  name        = "tfdrift-demo-web"
  description = "TFDrift demo - managed by Terraform, not by hand"
  vpc_id      = data.aws_vpc.demo.id

  # The only way in. Internal network only. This is the line the audience reads,
  # and the line the console change will contradict.
  ingress {
    description = "HTTPS from the internal network only"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    description = "all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name      = "tfdrift-demo-web"
    ManagedBy = "Terraform"
    Purpose   = "oss-summit-korea-2026-demo"
  }
}

output "security_group_id" {
  description = "Feed this to demo/drift-sg.sh as DEMO_SG"
  value       = aws_security_group.web.id
}
