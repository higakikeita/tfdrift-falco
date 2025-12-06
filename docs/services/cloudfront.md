# CloudFront — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateDistribution | Distribution created | ✔ |
| UpdateDistribution | Distribution updated | ✔ |
| DeleteDistribution | Distribution deleted | ✔ |
| CreateInvalidation | Cache invalidation requested | ✔ |
| UpdateOriginRequestPolicy | Origin request policy updated | ✔ |
| UpdateCachePolicy | Cache policy updated | ✔ |

## Monitored Drift Attributes

### Distribution
- enabled
- origins
  - domain_name
  - origin_path
  - custom_origin_config
    - http_port, https_port
    - origin_protocol_policy
    - origin_ssl_protocols
  - s3_origin_config
    - origin_access_identity
- default_cache_behavior
  - target_origin_id
  - viewer_protocol_policy (allow-all / https-only / redirect-to-https)
  - allowed_methods
  - cached_methods
  - compress
  - cache_policy_id
  - origin_request_policy_id
- viewer_certificate
  - acm_certificate_arn
  - minimum_protocol_version
  - ssl_support_method
- aliases (custom domain names)
- price_class
- geo_restriction
  - restriction_type (whitelist / blacklist / none)
  - locations

### Cache Policies
- name
- min_ttl, max_ttl, default_ttl
- parameters_in_cache_key_and_forwarded_to_origin
  - headers_config
  - cookies_config
  - query_strings_config

## Falco Rule Examples

```yaml
rule: cloudfront_https_disabled
condition:
  cloud.service = "cloudfront" and evt.name = "UpdateDistribution" and
  drift.changes.viewer_protocol_policy in ("allow-all","http-only")
output: "CloudFront HTTPS Disabled (distribution=%resource user=%user)"
priority: critical

rule: cloudfront_origin_changed
condition:
  cloud.service = "cloudfront" and evt.name = "UpdateDistribution" and
  drift.changes.origins != null
output: "CloudFront Origin Modified (distribution=%resource origins=%drift.changes.origins user=%user)"
priority: warning
```

## Example Log Output

```json
{
  "service": "cloudfront",
  "event": "UpdateDistribution",
  "resource": "E1234567890ABC",
  "changes": {
    "viewer_protocol_policy": ["redirect-to-https", "allow-all"],
    "origins": {
      "modified": [
        {
          "id": "S3-my-bucket",
          "domain_name": ["my-bucket.s3.amazonaws.com", "my-bucket-new.s3.amazonaws.com"]
        }
      ]
    }
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- CloudFront distribution updates
- HTTPS policy changes
- Origin modifications
- Cache policy updates

### Alerts
- HTTPS enforcement disabled
- Unplanned origin changes
- Geo-restriction removed
- SSL certificate changes

## Known Limitations

- CloudFront distribution updates have eventual consistency (may take 15-20 minutes)
- Real-time log configuration drift tracked separately
- Lambda@Edge function association drift partial (v0.3.0 planned)
- Field-level encryption drift not supported yet
- Origin Shield configuration changes partial

## Security Considerations

CloudFront drift detection is **critical for security and performance**:
- **HTTPS disabled** → man-in-the-middle attack risk
- **Origin changed** → content hijacking potential
- **Certificate changes** → SSL/TLS downgrade risk
- **Geo-restriction removed** → compliance violation

**Recommendation**: Set critical priority for viewer_protocol_policy and origin changes.

## Release History

- **v0.2.0-beta**: Core CloudFront distribution coverage (6 events)
- **v0.3.0** (planned): Lambda@Edge, Real-time Logs, Field-level Encryption
