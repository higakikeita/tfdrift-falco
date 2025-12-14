# ElastiCache (Amazon ElastiCache) — Drift Coverage

## Overview

TFDrift-Falco monitors Amazon ElastiCache for configuration drift by tracking CloudTrail events related to cache clusters, replication groups, and parameter groups. This enables real-time detection of manual changes made outside of Terraform workflows for both Redis and Memcached deployments.

## Supported CloudTrail Events

### Cache Cluster Management (4 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateCacheCluster | Cache cluster created | CRITICAL | ✔ |
| DeleteCacheCluster | Cache cluster deleted | CRITICAL | ✔ |
| ModifyCacheCluster | Cache cluster configuration modified | WARNING | ✔ |
| RebootCacheCluster | Cache cluster rebooted | WARNING | ✔ |

### Replication Group Management (5 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateReplicationGroup | Replication group created | CRITICAL | ✔ |
| DeleteReplicationGroup | Replication group deleted | CRITICAL | ✔ |
| ModifyReplicationGroup | Replication group configuration modified | WARNING | ✔ |
| IncreaseReplicaCount | Replica count increased | WARNING | ✔ |
| DecreaseReplicaCount | Replica count decreased | WARNING | ✔ |

### Parameter Group Management (3 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateCacheParameterGroup | Parameter group created | WARNING | ✔ |
| DeleteCacheParameterGroup | Parameter group deleted | WARNING | ✔ |
| ModifyCacheParameterGroup | Parameter group settings modified | HIGH | ✔ |

**Total: 12 CloudTrail events**

## Supported Terraform Resources

- `aws_elasticache_cluster` — ElastiCache cluster (Memcached or single-node Redis)
- `aws_elasticache_replication_group` — Redis replication group with automatic failover
- `aws_elasticache_parameter_group` — Cache parameter group configuration

## Monitored Drift Attributes

### ElastiCache Clusters
- **cluster_id** — Cluster identifier
- **engine** — Cache engine (redis, memcached)
- **engine_version** — Engine version (7.0, 6.2, etc.)
- **node_type** — Instance type (cache.t3.micro, cache.r6g.large, etc.)
- **num_cache_nodes** — Number of cache nodes
- **port** — Cache port number
- **parameter_group_name** — Associated parameter group
- **subnet_group_name** — Subnet group for VPC placement
- **security_group_ids** — Security group IDs
- **maintenance_window** — Maintenance window
- **snapshot_retention_limit** — Snapshot retention period (days)
- **snapshot_window** — Daily snapshot time window
- **az_mode** — Availability zone mode (single-az, cross-az)
- **preferred_availability_zones** — AZ placement
- **tags** — Resource tags

### Replication Groups
- **replication_group_id** — Replication group identifier
- **description** — Replication group description
- **engine** — Cache engine (redis only)
- **engine_version** — Redis version
- **node_type** — Instance type for all nodes
- **num_cache_clusters** — Number of cache clusters (nodes)
- **num_node_groups** — Number of node groups (shards)
- **replicas_per_node_group** — Read replicas per shard
- **automatic_failover_enabled** — Automatic failover configuration
- **multi_az_enabled** — Multi-AZ deployment
- **at_rest_encryption_enabled** — Encryption at rest
- **transit_encryption_enabled** — Encryption in transit
- **auth_token** — Authentication token (Redis AUTH)
- **kms_key_id** — KMS key for encryption
- **parameter_group_name** — Parameter group
- **subnet_group_name** — Subnet group
- **security_group_ids** — Security groups
- **maintenance_window** — Maintenance window
- **snapshot_retention_limit** — Snapshot retention
- **snapshot_window** — Snapshot time window
- **auto_minor_version_upgrade** — Auto version upgrades
- **tags** — Resource tags

### Parameter Groups
- **name** — Parameter group name
- **family** — Engine family (redis7, memcached1.6, etc.)
- **description** — Parameter group description
- **parameter** — Individual parameter configurations
- **tags** — Resource tags

## Falco Rule Examples

```yaml
# Cluster Lifecycle
- rule: ElastiCache Cluster Created
  desc: Detect when an ElastiCache cluster is created
  condition: >
    ct.name="CreateCacheCluster"
  output: >
    ElastiCache cluster created
    (user=%ct.user cluster=%ct.request.cacheClusterId engine=%ct.request.engine node_type=%ct.request.cacheNodeType
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, elasticache, cache]

- rule: ElastiCache Cluster Deleted
  desc: Detect when an ElastiCache cluster is deleted
  condition: >
    ct.name="DeleteCacheCluster"
  output: >
    ElastiCache cluster deleted
    (user=%ct.user cluster=%ct.request.cacheClusterId
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, elasticache, cache, security]

# Configuration Changes
- rule: ElastiCache Cluster Modified
  desc: Detect when an ElastiCache cluster configuration is modified
  condition: >
    ct.name="ModifyCacheCluster"
  output: >
    ElastiCache cluster modified
    (user=%ct.user cluster=%ct.request.cacheClusterId
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, elasticache, configuration]

# Replication Group Management
- rule: ElastiCache Replication Group Created
  desc: Detect when a replication group is created
  condition: >
    ct.name="CreateReplicationGroup"
  output: >
    ElastiCache replication group created
    (user=%ct.user group=%ct.request.replicationGroupId node_type=%ct.request.cacheNodeType
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, elasticache, redis, replication]

# Scaling Events
- rule: ElastiCache Replica Count Increased
  desc: Detect when replica count is increased
  condition: >
    ct.name="IncreaseReplicaCount"
  output: >
    ElastiCache replica count increased
    (user=%ct.user group=%ct.request.replicationGroupId new_count=%ct.request.newReplicaCount
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, elasticache, scaling]

# Parameter Changes
- rule: ElastiCache Parameter Group Modified
  desc: Detect when parameter group settings are modified
  condition: >
    ct.name="ModifyCacheParameterGroup"
  output: >
    ElastiCache parameter group modified
    (user=%ct.user group=%ct.request.cacheParameterGroupName
     region=%ct.region account=%ct.account)
  priority: HIGH
  source: aws_cloudtrail
  tags: [terraform, drift, elasticache, configuration]
```

## Example Drift Scenarios

### Scenario 1: Node Type Scaled Up Manually

**CloudTrail Event:**
```json
{
  "eventName": "ModifyCacheCluster",
  "requestParameters": {
    "cacheClusterId": "my-redis-cluster",
    "cacheNodeType": "cache.r6g.large",
    "applyImmediately": true
  },
  "userIdentity": {
    "principalId": "AIDAI23ABCD4EFGH5IJKL",
    "userName": "ops-engineer"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_elasticache_cluster.redis
Changed: node_type = cache.r6g.medium → cache.r6g.large
User: ops-engineer (IAM User)
Region: us-east-1
Severity: HIGH
```

### Scenario 2: Replication Group Deleted

**CloudTrail Event:**
```json
{
  "eventName": "DeleteReplicationGroup",
  "requestParameters": {
    "replicationGroupId": "prod-redis-cluster",
    "retainPrimaryCluster": false
  },
  "userIdentity": {
    "type": "AssumedRole",
    "principalId": "AROAI23ABCD4EFGH5IJKL:admin"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_elasticache_replication_group.prod_redis
DELETED: Replication group removed
User: admin (Assumed Role)
Region: us-east-1
Severity: CRITICAL
```

### Scenario 3: Replica Count Changed

**CloudTrail Event:**
```json
{
  "eventName": "IncreaseReplicaCount",
  "requestParameters": {
    "replicationGroupId": "redis-cache",
    "newReplicaCount": 3,
    "applyImmediately": true
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_elasticache_replication_group.cache
Changed: replicas_per_node_group = 2 → 3
User: sre-team (Assumed Role)
Region: us-east-1
Severity: MEDIUM
```

### Scenario 4: Parameter Group Modified

**CloudTrail Event:**
```json
{
  "eventName": "ModifyCacheParameterGroup",
  "requestParameters": {
    "cacheParameterGroupName": "redis7-params",
    "parameterNameValues": [
      {
        "parameterName": "maxmemory-policy",
        "parameterValue": "allkeys-lru"
      }
    ]
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_elasticache_parameter_group.redis7_params
Changed: parameter.maxmemory-policy = volatile-lru → allkeys-lru
User: developer@example.com (Console)
Region: us-east-1
Severity: HIGH
```

## Configuration Example

```yaml
# config.yaml
drift_rules:
  - name: "ElastiCache Cluster Configuration"
    resource_types:
      - "aws_elasticache_cluster"
    watched_attributes:
      - "node_type"
      - "num_cache_nodes"
      - "engine_version"
      - "parameter_group_name"
      - "security_group_ids"
    severity: "high"

  - name: "ElastiCache Replication Groups"
    resource_types:
      - "aws_elasticache_replication_group"
    watched_attributes:
      - "node_type"
      - "num_cache_clusters"
      - "replicas_per_node_group"
      - "automatic_failover_enabled"
      - "at_rest_encryption_enabled"
      - "transit_encryption_enabled"
    severity: "critical"

  - name: "ElastiCache Parameter Groups"
    resource_types:
      - "aws_elasticache_parameter_group"
    watched_attributes:
      - "parameter"
    severity: "high"
```

## Best Practices

### 1. Redis Replication Group with High Availability
```hcl
# Terraform - Multi-AZ Redis with automatic failover
resource "aws_elasticache_replication_group" "redis" {
  replication_group_id       = "prod-redis"
  description                = "Production Redis cluster"
  engine                     = "redis"
  engine_version             = "7.0"
  node_type                  = "cache.r6g.large"
  num_cache_clusters         = 3
  port                       = 6379
  parameter_group_name       = aws_elasticache_parameter_group.redis7.name
  subnet_group_name          = aws_elasticache_subnet_group.redis.name
  security_group_ids         = [aws_security_group.redis.id]

  # High availability
  automatic_failover_enabled = true
  multi_az_enabled           = true

  # Security
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = var.redis_auth_token
  kms_key_id                 = aws_kms_key.elasticache.arn

  # Backup
  snapshot_retention_limit   = 7
  snapshot_window            = "03:00-04:00"
  maintenance_window         = "sun:04:00-sun:05:00"

  # Version management
  auto_minor_version_upgrade = true

  tags = {
    Name        = "prod-redis"
    Environment = "production"
  }
}
```

### 2. Memcached Cluster
```hcl
# Terraform - Memcached cluster for session storage
resource "aws_elasticache_cluster" "memcached" {
  cluster_id           = "session-cache"
  engine               = "memcached"
  engine_version       = "1.6.17"
  node_type            = "cache.t3.medium"
  num_cache_nodes      = 3
  parameter_group_name = "default.memcached1.6"
  port                 = 11211
  subnet_group_name    = aws_elasticache_subnet_group.cache.name
  security_group_ids   = [aws_security_group.memcached.id]

  # Availability
  az_mode                      = "cross-az"
  preferred_availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]

  # Maintenance
  maintenance_window = "sun:05:00-sun:06:00"

  tags = {
    Name        = "session-cache"
    Environment = "production"
  }
}
```

### 3. Parameter Group for Redis
```hcl
# Terraform - Custom Redis parameter group
resource "aws_elasticache_parameter_group" "redis7" {
  name   = "redis7-custom"
  family = "redis7"

  # Memory management
  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }

  parameter {
    name  = "maxmemory-samples"
    value = "10"
  }

  # Persistence
  parameter {
    name  = "appendonly"
    value = "yes"
  }

  parameter {
    name  = "appendfsync"
    value = "everysec"
  }

  # Timeouts
  parameter {
    name  = "timeout"
    value = "300"
  }

  # Connection limits
  parameter {
    name  = "maxclients"
    value = "10000"
  }

  tags = {
    Name = "redis7-custom"
  }
}
```

### 4. Subnet Group
```hcl
# Terraform - Subnet group for VPC placement
resource "aws_elasticache_subnet_group" "redis" {
  name       = "redis-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "redis-subnet-group"
  }
}
```

## Security Considerations

### 1. Encryption
- Enable at-rest encryption for sensitive data
- Enable transit encryption (TLS) for all connections
- Monitor encryption configuration changes
- Use AWS KMS for key management

### 2. Authentication
- Use Redis AUTH token for Redis clusters
- Rotate auth tokens regularly
- Store auth tokens in AWS Secrets Manager
- Monitor auth token changes

### 3. Network Security
- Deploy in private subnets only
- Use security groups to restrict access
- Monitor security group changes
- Never expose ElastiCache publicly

### 4. Parameter Security
- Review parameter group changes carefully
- Monitor maxclients and maxmemory settings
- Audit timeout and connection parameters
- Track changes to persistence settings

### 5. Backup and Recovery
- Configure automatic snapshots
- Set appropriate retention periods
- Test restore procedures regularly
- Monitor snapshot window changes

## Known Limitations

### 1. Cluster vs Replication Group Context
- Some operations apply to both resource types
- ModifyCluster can apply to standalone or grouped clusters
- Use resource tagging for clear identification

### 2. Parameter Change Details
- ModifyCacheParameterGroup tracks changes but may not show all parameters
- Compare Terraform state for detailed parameter drift
- Some parameters require reboot to take effect

### 3. Scaling Operations
- Node type changes may require cluster recreation
- Scaling operations may take several minutes
- In-flight scaling not distinguished from configuration drift

### 4. Engine Version Updates
- Minor version upgrades tracked via ModifyReplicationGroup
- Major version upgrades require careful planning
- Engine version format variations (7.0 vs 7.0.7)

## Related Documentation

- [AWS ElastiCache CloudTrail Logging](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/logging-using-cloudtrail.html)
- [Terraform aws_elasticache_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elasticache_cluster)
- [Terraform aws_elasticache_replication_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elasticache_replication_group)
- [Terraform aws_elasticache_parameter_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elasticache_parameter_group)
- [Terraform ElastiCache Module](https://registry.terraform.io/modules/terraform-aws-modules/elasticache/aws/latest)

## Version History

- **v0.3.0** (2025 Q1) - Initial ElastiCache support with 12 CloudTrail events
  - Cache cluster management (create, delete, modify, reboot)
  - Replication group management (create, delete, modify, scale)
  - Parameter group management (create, delete, modify)
  - Comprehensive drift detection for Redis and Memcached caching infrastructure
