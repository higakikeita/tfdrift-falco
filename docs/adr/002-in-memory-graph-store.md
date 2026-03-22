# ADR-002: In-Memory Graph Store Over External Database

## Status

Accepted

## Date

2026-03-22

## Context

TFDrift-Falco builds a dependency graph of cloud resources, Terraform state, and drift relationships. This graph needs to support queries like "show all resources affected by this drift" and "visualize the topology of resources in this VPC."

We considered several storage options:

1. **External graph database** (Neo4j, Amazon Neptune) — Full-featured graph queries
2. **External relational database** (PostgreSQL with recursive CTEs) — Familiar, widely deployed
3. **In-memory graph store** — Custom Go implementation with adjacency lists
4. **Embedded database** (BoltDB, BadgerDB) — Persistent, no external dependency

## Decision

We use a custom in-memory graph store (`pkg/graph/store.go`) backed by Go maps and adjacency lists. The store provides node/edge CRUD, neighbor traversal, subgraph extraction, and export capabilities (PNG, SVG, JSON).

## Consequences

### Positive

- Zero external dependencies — no database to deploy or manage
- Sub-millisecond query performance for typical graph sizes (< 10,000 nodes)
- Simple deployment — single binary, no connection strings or migrations
- Full control over query patterns optimized for our use cases
- Export to multiple formats directly from memory

### Negative

- Data is lost on restart (no persistence)
- Memory-bound — graph size limited by available RAM
- No built-in Cypher/Gremlin query language — custom query API required
- Scaling limited to single instance (no clustering)
- Must implement our own concurrency control (sync.RWMutex)

### Neutral

- Suitable for the current scale (< 50,000 resources per instance)
- A future ADR may supersede this if persistence or scale requirements change
- The graph store interface is abstract enough to swap implementations later
