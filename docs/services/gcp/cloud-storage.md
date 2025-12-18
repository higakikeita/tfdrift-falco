# GCP Cloud Storage

> **Service:** Google Cloud Storage (GCS)
> **Events Monitored:** 5+
> **Resources:** `google_storage_bucket`, `google_storage_bucket_object`, `google_storage_bucket_iam_binding`
> **Status:** ✅ Production Ready

## Monitored Events

| Event Name | Description | Resource Type |
|------------|-------------|---------------|
| `storage.buckets.create` | Bucket creation | `google_storage_bucket` |
| `storage.buckets.delete` | Bucket deletion | `google_storage_bucket` |
| `storage.buckets.update` | Bucket configuration update | `google_storage_bucket` |
| `storage.buckets.patch` | Bucket patch update | `google_storage_bucket` |
| `storage.buckets.setIamPolicy` | IAM policy change | `google_storage_bucket_iam_binding` |
| `storage.objects.create` | Object creation | `google_storage_bucket_object` |
| `storage.objects.delete` | Object deletion | `google_storage_bucket_object` |

## Example Configuration

```yaml
drift_rules:
  - name: "GCS Bucket IAM Change"
    resource_types:
      - "google_storage_bucket_iam_binding"
    watched_attributes:
      - "members"
    severity: "critical"

  - name: "GCS Bucket Public Access"
    resource_types:
      - "google_storage_bucket"
    watched_attributes:
      - "iam_configuration"
    severity: "critical"
```

## Documentation

- [GCP Services Overview](index.md)
- [GCS Documentation](https://cloud.google.com/storage/docs)
- [Terraform google_storage_bucket](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_bucket)
