# VPC — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateVpc | VPC created | ✔ |
| DeleteVpc | VPC deleted | ✔ |
| CreateSubnet | Subnet created | ✔ |
| DeleteSubnet | Subnet deleted | ✔ |
| ModifySubnetAttribute | Subnet attribute modified | ✔ |
| CreateSecurityGroup | Security group created | ✔ |
| DeleteSecurityGroup | Security group deleted | ✔ |
| AuthorizeSecurityGroupIngress | Ingress rule added | ✔ |
| AuthorizeSecurityGroupEgress | Egress rule added | ✔ |
| RevokeSecurityGroupIngress | Ingress rule removed | ✔ |
| RevokeSecurityGroupEgress | Egress rule removed | ✔ |
| CreateRouteTable | Route table created | ✔ |
| CreateRoute | Route added | ✔ |
| DeleteRoute | Route deleted | ✔ |
| AssociateRouteTable | Route table associated | ✔ |
| CreateInternetGateway | IGW created | ✔ |
| AttachInternetGateway | IGW attached to VPC | ✔ |
| CreateNatGateway | NAT Gateway created | ✔ |
| DeleteNatGateway | NAT Gateway deleted | ✔ |

## Monitored Drift Attributes

### VPC
- cidr_block
- enable_dns_hostnames
- enable_dns_support
- instance_tenancy
- tags

### Subnet
- cidr_block
- availability_zone
- map_public_ip_on_launch
- tags

### Security Group
- name
- description
- vpc_id
- ingress rules
  - from_port, to_port, protocol
  - cidr_blocks, ipv6_cidr_blocks
  - source_security_group_id
- egress rules

### Route Table
- routes
  - destination_cidr_block
  - gateway_id, nat_gateway_id, instance_id, vpc_peering_connection_id
- subnet associations

## Falco Rule Examples

```yaml
rule: security_group_ingress_0_0_0_0
condition:
  cloud.service = "ec2" and evt.name = "AuthorizeSecurityGroupIngress" and
  drift.cidr_blocks contains "0.0.0.0/0"
output: "Security Group Opened to Internet (sg=%resource port=%drift.from_port-%drift.to_port user=%user)"
priority: critical

rule: route_table_modified
condition:
  cloud.service = "ec2" and evt.name in ("CreateRoute","DeleteRoute") and
  drift.planned = false
output: "Unplanned Route Table Change (table=%resource destination=%drift.destination_cidr_block user=%user)"
priority: warning
```

## Example Log Output

```json
{
  "service": "ec2",
  "event": "AuthorizeSecurityGroupIngress",
  "resource": "sg-0123456789abcdef0",
  "changes": {
    "ingress_added": [
      {
        "from_port": 22,
        "to_port": 22,
        "protocol": "tcp",
        "cidr_blocks": ["0.0.0.0/0"]
      }
    ]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- Security group rule changes by group
- Route table modifications
- NAT Gateway creations/deletions
- Subnet public IP assignment changes

### Alerts
- 0.0.0.0/0 ingress rules added
- Unplanned route deletions
- VPC peering changes
- NAT Gateway deletions

## Known Limitations

- VPC Flow Logs configuration not tracked (separate service)
- Transit Gateway attachment drift partial (v0.3.0 planned)
- VPC Endpoint policy changes tracked but service-specific policies not parsed
- Network ACL drift tracked but priority evaluation not analyzed

## Security Considerations

VPC drift detection is **critical for network security**:
- **0.0.0.0/0 ingress** → potential breach
- **Route table changes** → traffic redirection risk
- **Security group deletions** → service disruption

**Recommendation**: Set critical priority for security group rules with 0.0.0.0/0.

## Release History

- **v0.2.0-beta**: Core VPC/subnet/security group/route table coverage (19 events)
- **v0.3.0** (planned): Transit Gateway, VPC Endpoint advanced features
