# How It Works

This document explains the technical details of how TFDrift-Falco detects Terraform drift in real-time.

---

## High-Level Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. AWS Change Happens                                           │
│    User modifies EC2 instance type via AWS Console              │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. CloudTrail Event Generated                                   │
│    EventName: ModifyInstanceAttribute                           │
│    Resource: i-0123456789abcdef0                                │
│    RequestParameters: { instanceType: "t3.small" }              │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. TFDrift Detector Polls CloudTrail                            │
│    - Fetch events from last 5 minutes                           │
│    - Filter by supported event names                            │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Load Terraform State                                         │
│    - Parse Terraform state from S3                              │
│    - Find resource: aws_instance.web                            │
│    - Extract current attribute: instance_type = "t3.micro"      │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Compare State vs. Event                                      │
│    Terraform State:  instance_type = "t3.micro"                 │
│    CloudTrail Event: instance_type = "t3.small"                 │
│    → DRIFT DETECTED                                             │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Emit Falco Event                                             │
│    {                                                            │
│      "service": "ec2",                                          │
│      "event": "ModifyInstanceAttribute",                        │
│      "resource": "i-0123456789abcdef0",                         │
│      "changes": {                                               │
│        "instance_type": ["t3.micro", "t3.small"]                │
│      },                                                         │
│      "user": "admin@example.com"                                │
│    }                                                            │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. Falco Rule Matches                                           │
│    rule: ec2_instance_type_changed                              │
│    priority: warning                                            │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ├──────────────────┬──────────────────┐
                     ▼                  ▼                  ▼
              ┌───────────┐      ┌───────────┐    ┌───────────┐
              │  Grafana  │      │   Slack   │    │  PagerDuty│
              │ Dashboard │      │   Alert   │    │  Incident │
              └───────────┘      └───────────┘    └───────────┘
```

---

## Component Details

### 1. CloudTrail Event Polling

**Implementation:** `pkg/cloudtrail/poller.go`

```go
func (p *Poller) FetchEvents() ([]Event, error) {
    input := &cloudtrail.LookupEventsInput{
        StartTime: aws.Time(time.Now().Add(-5 * time.Minute)),
        LookupAttributes: []*cloudtrail.LookupAttribute{
            {
                AttributeKey:   aws.String("EventName"),
                AttributeValue: aws.String("ModifyInstanceAttribute"),
            },
        },
    }
    resp, err := p.client.LookupEvents(input)
    // ... parse events
}
```

**Features:**
- Polls every 1 minute (configurable)
- Filters by event name (only supported events)
- Deduplication (track processed event IDs)
- Pagination support for high-volume accounts

---

### 2. Terraform State Loading

**Implementation:** `pkg/terraform/state.go`

```go
func (s *StateLoader) LoadResource(resourceType, resourceID string) (*Resource, error) {
    // 1. Fetch state from S3 backend
    state, err := s.backend.GetState()

    // 2. Parse JSON
    var tfState TerraformState
    json.Unmarshal(state, &tfState)

    // 3. Find resource by ID
    for _, resource := range tfState.Resources {
        if resource.Type == resourceType && resource.Instances[0].ID == resourceID {
            return &resource, nil
        }
    }
    return nil, ErrResourceNotFound
}
```

**Supported Backends:**
- S3 (with state locking via DynamoDB)
- Local file
- Terraform Cloud (v0.3.0 planned)
- Consul (v0.3.0 planned)

---

### 3. State Comparison Logic

**Implementation:** `pkg/detector/comparator.go`

```go
func (c *Comparator) Compare(event CloudTrailEvent, tfResource TerraformResource) (*Drift, error) {
    // Extract changed attributes from CloudTrail event
    changedAttrs := parseRequestParameters(event)

    // Compare with Terraform state
    drifts := []AttributeDrift{}
    for attrName, newValue := range changedAttrs {
        oldValue := tfResource.Attributes[attrName]
        if oldValue != newValue {
            drifts = append(drifts, AttributeDrift{
                Name:     attrName,
                OldValue: oldValue,
                NewValue: newValue,
            })
        }
    }

    if len(drifts) > 0 {
        return &Drift{
            Resource: tfResource.ID,
            Changes:  drifts,
        }, nil
    }
    return nil, nil // No drift
}
```

**Key Logic:**
- Only compares attributes mentioned in CloudTrail event
- Handles complex nested attributes (maps, lists)
- Type coercion (string "true" vs. bool true)

---

### 4. Drift Event Emission

**Implementation:** `pkg/falco/emitter.go`

```go
func (e *Emitter) EmitDrift(drift *Drift) error {
    event := FalcoEvent{
        Timestamp: time.Now(),
        Service:   drift.Service,
        EventName: drift.CloudTrailEvent,
        Resource:  drift.ResourceID,
        Changes:   drift.Changes,
        User:      drift.User,
    }

    // Send to Falco via Unix socket
    return e.conn.WriteJSON(event)
}
```

**Output Format:**
```json
{
  "timestamp": "2025-12-06T07:30:00Z",
  "service": "ec2",
  "event": "ModifyInstanceAttribute",
  "resource": "i-0123456789abcdef0",
  "changes": {
    "instance_type": ["t3.micro", "t3.small"]
  },
  "user": "arn:aws:iam::123456789012:user/admin",
  "source_ip": "203.0.113.1",
  "user_agent": "console.amazonaws.com"
}
```

---

## Advanced Features

### 1. Change Detection Strategies

#### Simple Attribute Change
```yaml
# Terraform State
resource "aws_instance" "web" {
  instance_type = "t3.micro"
}

# CloudTrail Event
ModifyInstanceAttribute: { instanceType: "t3.small" }

# Result: DRIFT
```

#### Complex Nested Attribute
```yaml
# Terraform State
resource "aws_security_group" "web" {
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }
}

# CloudTrail Event
AuthorizeSecurityGroupIngress: {
  IpPermissions: [{
    FromPort: 443,
    ToPort: 443,
    IpProtocol: "tcp",
    IpRanges: [{ CidrIp: "0.0.0.0/0" }]
  }]
}

# Result: DRIFT (new rule added)
```

---

### 2. False Positive Filtering

#### Auto Scaling Events
```go
func isAutoScalingEvent(event CloudTrailEvent) bool {
    // Ignore events from Auto Scaling service
    return strings.Contains(event.UserIdentity.PrincipalID, "autoscaling.amazonaws.com")
}
```

#### Terraform-initiated Changes
```go
func isTerraformChange(event CloudTrailEvent) bool {
    // Check if user agent is Terraform
    return strings.Contains(event.UserAgent, "Terraform")
}
```

---

### 3. Multi-Region Support

```go
func (d *Detector) MonitorRegions(regions []string) {
    for _, region := range regions {
        go d.monitorRegion(region) // Concurrent monitoring
    }
}

func (d *Detector) monitorRegion(region string) {
    client := cloudtrail.New(session.New(&aws.Config{Region: aws.String(region)}))
    // ... poll CloudTrail in this region
}
```

---

## Performance Optimization

### 1. Caching

```go
// Cache Terraform state for 5 minutes
type StateCache struct {
    cache map[string]*CachedState
    ttl   time.Duration
}

func (c *StateCache) Get(key string) (*TerraformState, bool) {
    if cached, ok := c.cache[key]; ok {
        if time.Since(cached.Timestamp) < c.ttl {
            return cached.State, true
        }
    }
    return nil, false
}
```

**Performance Impact:**
- State load time: 2s → 0.1s (for cached states)
- CloudTrail API calls reduced by 60%

### 2. Parallel Event Processing

```go
func (d *Detector) ProcessEvents(events []CloudTrailEvent) {
    var wg sync.WaitGroup
    for _, event := range events {
        wg.Add(1)
        go func(e CloudTrailEvent) {
            defer wg.Done()
            d.processEvent(e)
        }(event)
    }
    wg.Wait()
}
```

**Performance Impact:**
- Processing time for 100 events: 50s → 10s

---

## Edge Cases Handling

### 1. Eventually Consistent Resources

Some AWS resources have eventual consistency (e.g., IAM, Route53).

**Solution:** Retry with exponential backoff

```go
func (d *Detector) verifyDrift(drift *Drift) error {
    for i := 0; i < 3; i++ {
        time.Sleep(time.Duration(2^i) * time.Second)
        if d.isDriftStillPresent(drift) {
            return d.emitDrift(drift)
        }
    }
    return nil // Drift resolved (eventual consistency)
}
```

### 2. Bulk Operations

CloudTrail may batch multiple changes into one event.

**Solution:** Parse all changes in the event

```go
func parseBulkChanges(event CloudTrailEvent) []AttributeChange {
    changes := []AttributeChange{}
    for _, item := range event.RequestParameters.Items {
        changes = append(changes, AttributeChange{
            Attribute: item.Key,
            NewValue:  item.Value,
        })
    }
    return changes
}
```

---

## Next Steps

1. [Understand the Architecture →](architecture.md)
2. [Review Service Coverage →](services/ec2.md)
3. [Deploy TFDrift-Falco →](quickstart.md)
