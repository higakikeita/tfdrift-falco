#!/bin/bash
#
# TFDrift-Falco GCP Quick Start Script
#
# This script automates the setup of TFDrift-Falco for Google Cloud Platform.
# It creates the necessary GCP resources, installs Falco, and configures TFDrift-Falco.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/higakikeita/tfdrift-falco/main/scripts/gcp-quick-start.sh | bash
#
# Or download and run:
#   chmod +x gcp-quick-start.sh
#   ./gcp-quick-start.sh
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_step() {
    echo -e "${BLUE}==>${NC} ${1}"
}

print_success() {
    echo -e "${GREEN}✓${NC} ${1}"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} ${1}"
}

print_error() {
    echo -e "${RED}✗${NC} ${1}"
}

check_prerequisites() {
    print_step "Checking prerequisites..."

    # Check gcloud
    if ! command -v gcloud &> /dev/null; then
        print_error "gcloud CLI is not installed. Please install it first:"
        echo "  https://cloud.google.com/sdk/docs/install"
        exit 1
    fi
    print_success "gcloud CLI found"

    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install it first:"
        echo "  https://docs.docker.com/get-docker/"
        exit 1
    fi
    print_success "Docker found"

    # Check if Docker is running
    if ! docker ps &> /dev/null; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    print_success "Docker is running"

    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        print_warning "Terraform is not installed. You'll need it to manage GCP resources."
        echo "  Install from: https://www.terraform.io/downloads"
    else
        print_success "Terraform found"
    fi
}

get_project_id() {
    print_step "Getting GCP project..."

    PROJECT_ID=$(gcloud config get-value project 2>/dev/null)

    if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "(unset)" ]; then
        print_error "No GCP project is set. Please set one:"
        echo "  gcloud config set project YOUR_PROJECT_ID"
        exit 1
    fi

    print_success "Using project: $PROJECT_ID"
}

enable_apis() {
    print_step "Enabling required GCP APIs..."

    gcloud services enable \
        logging.googleapis.com \
        pubsub.googleapis.com \
        compute.googleapis.com \
        storage-api.googleapis.com \
        --project="$PROJECT_ID" \
        --quiet

    print_success "APIs enabled"
}

create_pubsub() {
    print_step "Creating Pub/Sub infrastructure..."

    # Create topic
    if gcloud pubsub topics describe tfdrift-audit-logs --project="$PROJECT_ID" &>/dev/null; then
        print_warning "Pub/Sub topic 'tfdrift-audit-logs' already exists"
    else
        gcloud pubsub topics create tfdrift-audit-logs \
            --project="$PROJECT_ID" \
            --quiet
        print_success "Created Pub/Sub topic"
    fi

    # Create log sink
    if gcloud logging sinks describe tfdrift-sink --project="$PROJECT_ID" &>/dev/null; then
        print_warning "Log sink 'tfdrift-sink' already exists"
    else
        gcloud logging sinks create tfdrift-sink \
            pubsub.googleapis.com/projects/$PROJECT_ID/topics/tfdrift-audit-logs \
            --log-filter='protoPayload.serviceName="compute.googleapis.com"' \
            --project="$PROJECT_ID" \
            --quiet
        print_success "Created log sink"
    fi

    # Grant permissions to sink
    SINK_SA=$(gcloud logging sinks describe tfdrift-sink --project="$PROJECT_ID" --format="value(writerIdentity)")
    gcloud pubsub topics add-iam-policy-binding tfdrift-audit-logs \
        --member="$SINK_SA" \
        --role="roles/pubsub.publisher" \
        --project="$PROJECT_ID" \
        --quiet
    print_success "Granted permissions to log sink"

    # Create subscription
    if gcloud pubsub subscriptions describe tfdrift-falco-sub --project="$PROJECT_ID" &>/dev/null; then
        print_warning "Subscription 'tfdrift-falco-sub' already exists"
    else
        gcloud pubsub subscriptions create tfdrift-falco-sub \
            --topic=tfdrift-audit-logs \
            --project="$PROJECT_ID" \
            --quiet
        print_success "Created Pub/Sub subscription"
    fi
}

create_service_account() {
    print_step "Creating service account for Falco..."

    SA_EMAIL="tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com"

    # Check if service account exists
    if gcloud iam service-accounts describe $SA_EMAIL --project="$PROJECT_ID" &>/dev/null; then
        print_warning "Service account already exists"
    else
        gcloud iam service-accounts create tfdrift-falco \
            --display-name="TFDrift Falco Service Account" \
            --project="$PROJECT_ID" \
            --quiet
        print_success "Created service account"
    fi

    # Grant permissions
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$SA_EMAIL" \
        --role="roles/pubsub.subscriber" \
        --quiet
    print_success "Granted Pub/Sub subscriber role"

    # Create key
    mkdir -p ~/tfdrift-config
    if [ -f ~/tfdrift-config/gcp-key.json ]; then
        print_warning "Service account key already exists at ~/tfdrift-config/gcp-key.json"
    else
        gcloud iam service-accounts keys create ~/tfdrift-config/gcp-key.json \
            --iam-account=$SA_EMAIL \
            --project="$PROJECT_ID" \
            --quiet
        print_success "Created service account key: ~/tfdrift-config/gcp-key.json"
    fi
}

configure_falco() {
    print_step "Configuring Falco..."

    cat > ~/tfdrift-config/falco.yaml <<EOF
# Falco configuration for TFDrift-Falco (GCP)
# Generated by gcp-quick-start.sh

# Use modern eBPF engine (no kernel module needed for cloud audit logs)
engine:
  kind: modern_ebpf
  modern_ebpf:
    cpus_for_each_buffer: 2

# Load GCP audit plugin
plugins:
  - name: gcpaudit
    library_path: /usr/share/falco/plugins/libgcpaudit.so
    init_config:
      project_id: "$PROJECT_ID"
      subscription: "tfdrift-falco-sub"
    open_params: ""

# Load rules for GCP
load_plugins: [gcpaudit]

# Output configuration
json_output: true
json_include_output_property: true

# gRPC server configuration
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 8

grpc_output:
  enabled: true
EOF

    print_success "Created Falco configuration: ~/tfdrift-config/falco.yaml"
}

run_falco() {
    print_step "Starting Falco..."

    # Stop existing container if running
    if docker ps -a | grep -q falco; then
        print_warning "Stopping existing Falco container..."
        docker stop falco &>/dev/null || true
        docker rm falco &>/dev/null || true
    fi

    # Run Falco
    docker run -d \
        --name falco \
        -p 5060:5060 \
        -v ~/tfdrift-config:/etc/falco \
        -e GOOGLE_APPLICATION_CREDENTIALS=/etc/falco/gcp-key.json \
        falcosecurity/falco:latest \
        -c /etc/falco/falco.yaml

    print_success "Falco is running (container: falco)"

    # Wait for Falco to start
    print_step "Waiting for Falco to initialize..."
    sleep 5

    # Check if Falco is running
    if docker ps | grep -q falco; then
        print_success "Falco is ready"
    else
        print_error "Falco failed to start. Check logs with: docker logs falco"
        exit 1
    fi
}

configure_tfdrift() {
    print_step "Configuring TFDrift-Falco..."

    cat > ~/tfdrift-config/config-gcp.yaml <<EOF
# TFDrift-Falco Configuration (GCP)
# Generated by gcp-quick-start.sh

providers:
  gcp:
    enabled: true
    projects:
      - "$PROJECT_ID"
    state:
      backend: "local"
      local_path: "./terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  - name: "GCE Instance Configuration Change"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
      - "tags"
      - "machine_type"
    severity: "high"

  - name: "Firewall Rule Modification"
    resource_types:
      - "google_compute_firewall"
    watched_attributes:
      - "allow"
      - "deny"
      - "source_ranges"
    severity: "critical"

notifications:
  slack:
    enabled: false
    # webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

logging:
  level: "info"
  format: "text"
EOF

    print_success "Created TFDrift-Falco configuration: ~/tfdrift-config/config-gcp.yaml"
}

print_summary() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Setup Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Configuration files created in: ~/tfdrift-config/"
    echo "  - falco.yaml"
    echo "  - config-gcp.yaml"
    echo "  - gcp-key.json"
    echo ""
    echo "Falco container is running on port 5060"
    echo ""
    echo -e "${BLUE}Next Steps:${NC}"
    echo ""
    echo "1. Create a test Terraform resource:"
    echo "   cat > main.tf <<'TFEOF'"
    echo "   resource \"google_compute_network\" \"test\" {"
    echo "     name                    = \"tfdrift-test-network\""
    echo "     auto_create_subnetworks = false"
    echo "   }"
    echo "   TFEOF"
    echo "   terraform init && terraform apply -auto-approve"
    echo ""
    echo "2. Run TFDrift-Falco:"
    echo "   tfdrift --config ~/tfdrift-config/config-gcp.yaml"
    echo ""
    echo "3. Make a manual change to trigger drift detection:"
    echo "   gcloud compute networks update tfdrift-test-network \\"
    echo "     --description=\"Manual change - should trigger drift\""
    echo ""
    echo "4. Check Falco logs:"
    echo "   docker logs -f falco"
    echo ""
    echo -e "${BLUE}Clean Up (when done testing):${NC}"
    echo "  terraform destroy -auto-approve"
    echo "  docker stop falco && docker rm falco"
    echo "  gcloud pubsub subscriptions delete tfdrift-falco-sub --project=$PROJECT_ID"
    echo "  gcloud pubsub topics delete tfdrift-audit-logs --project=$PROJECT_ID"
    echo "  gcloud logging sinks delete tfdrift-sink --project=$PROJECT_ID"
    echo ""
    echo -e "${YELLOW}Documentation:${NC}"
    echo "  https://github.com/higakikeita/tfdrift-falco/blob/main/docs/gcp-setup.md"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}TFDrift-Falco GCP Quick Start${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    check_prerequisites
    get_project_id
    enable_apis
    create_pubsub
    create_service_account
    configure_falco
    run_falco
    configure_tfdrift
    print_summary
}

# Run main function
main
