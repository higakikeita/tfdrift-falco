# ADR-003: Multi-Cloud Provider Abstraction Pattern

## Status

Accepted

## Date

2026-03-22

## Context

TFDrift-Falco supports drift detection across AWS, GCP, and Azure. Each cloud provider has different audit log formats, API naming conventions, and Terraform resource type mappings. We needed a way to handle these differences without duplicating core logic.

Options considered:

1. **Single monolithic parser** — Handle all providers in one package with switch statements
2. **Provider-specific packages** — Separate packages per provider with a common interface
3. **Plugin architecture** — Dynamically loaded provider plugins

## Decision

We adopt provider-specific packages (`pkg/falco/` for AWS, `pkg/gcp/` for GCP, `pkg/azure/` for Azure) that share a common event type (`pkg/types/types.go`). Each provider has:

- A **parser** that converts raw audit log events into the common `Event` type
- A **resource mapper** that maps cloud API operations to Terraform resource types
- A **mapping table** that defines the event-to-resource-type relationships

The Falco subscriber routes events to the appropriate parser based on the event source field (`aws_cloudtrail`, `gcpaudit`, `azureaudit`).

## Consequences

### Positive

- Clean separation of concerns — each provider is independent
- Easy to add new providers without modifying existing code
- Provider-specific logic is contained and testable in isolation
- Common event type enables provider-agnostic downstream processing (drift detection, graph building, API responses)
- Mapping tables are declarative and easy to extend

### Negative

- Some code duplication across providers (parser structure, mapping table format)
- Adding a new provider requires understanding the pattern and creating multiple files
- The common `Event` type must accommodate fields from all providers, leading to some optional fields

### Neutral

- AWS has the most mature mapping table (500+ events, 40+ services); GCP and Azure are catching up
- Conflict resolution logic (`pkg/falco/mappings/conflicts.go`) handles overlapping event names across AWS services
