# API Gateway — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateRestApi | REST API created | ✔ |
| DeleteRestApi | REST API deleted | ✔ |
| UpdateRestApi | REST API updated | ✔ |
| CreateDeployment | New deployment created | ✔ |
| UpdateStage | Stage configuration updated | ✔ |
| CreateAuthorizer | Authorizer added | ✔ |
| UpdateAuthorizer | Authorizer updated | ✔ |
| DeleteAuthorizer | Authorizer removed | ✔ |
| PutRestApi | API imported/updated via OpenAPI | ✔ |

## Monitored Drift Attributes

### REST / HTTP API
- name
- description
- endpoint_type (EDGE / REGIONAL / PRIVATE)
- api_key_source (HEADER / AUTHORIZER)
- minimum_compression_size
- binary_media_types
- disable_execute_api_endpoint

### Stages
- stage_name
- deployment_id
- access_log_settings
  - destination_arn (CloudWatch Logs)
  - format
- throttling_burst_limit
- throttling_rate_limit
- cache_cluster_enabled
- cache_cluster_size
- variables

### Authorizers
- type (TOKEN / REQUEST / COGNITO_USER_POOLS)
- identity_source (e.g., method.request.header.Authorization)
- authorizer_uri (Lambda function ARN)
- authorizer_credentials
- authorizer_result_ttl_in_seconds

## Falco Rule Examples

```yaml
rule: apigw_authorizer_modified
condition:
  cloud.service = "apigateway" and evt.name in ("UpdateAuthorizer","DeleteAuthorizer")
output: "API Gateway Authorizer Changed (api=%resource authorizer=%drift.authorizer_name changes=%drift.changes user=%user)"
priority: critical

rule: apigw_throttling_disabled
condition:
  cloud.service = "apigateway" and evt.name = "UpdateStage" and
  drift.changes.throttling_rate_limit = null
output: "API Gateway Throttling Disabled (api=%resource stage=%drift.stage_name user=%user)"
priority: warning
```

## Example Log Output

```json
{
  "service": "api-gateway",
  "event": "UpdateAuthorizer",
  "resource": "my-api",
  "changes": {
    "authorizer_name": "my-auth",
    "identity_source": [
      "method.request.header.Authorization",
      "method.request.header.X-API-Key"
    ],
    "authorizer_result_ttl_in_seconds": [300, 0]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- API Gateway authorizer changes
- Stage deployment frequency
- Throttling configuration changes
- Access logging modifications

### Alerts
- Unplanned authorizer deletions
- Throttling disabled
- Access logging removed
- Private API exposure

## Known Limitations

- WebSocket API drift is partial (v0.3.0 planned for full coverage)
- Import/export via OpenAPI definition drift not fully parsed (diff detection only)
- Custom domain name drift tracked separately
- API Gateway v1 vs v2 (HTTP API) have different event structures
- Request/response transformation drift not analyzed

## Security Considerations

API Gateway drift detection is **critical for API security**:
- **Authorizer removal** → authentication bypass
- **Throttling disabled** → DDoS vulnerability
- **Access logging removed** → audit trail loss
- **Private API made public** → data exposure

**Recommendation**: Set critical priority for authorizer and endpoint type changes.

## Release History

- **v0.2.0-beta**: REST API core coverage (9 events)
- **v0.3.0** (planned): WebSocket API, HTTP API v2 enhancements
