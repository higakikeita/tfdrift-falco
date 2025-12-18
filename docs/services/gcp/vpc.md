# GCP VPC & Firewall

> **Service:** VPC Networking & Firewall Rules
> **Events Monitored:** 11+
> **Resources:** `google_compute_network`, `google_compute_subnetwork`, `google_compute_firewall`
> **Status:** ✅ Production Ready

## Monitored Events

### Firewall Rules (4 events)
- `compute.firewalls.insert` - Firewall rule creation
- `compute.firewalls.delete` - Firewall rule deletion
- `compute.firewalls.update` - Firewall rule update
- `compute.firewalls.patch` - Firewall rule patch

### Networks (3 events)
- `compute.networks.insert` - Network creation
- `compute.networks.delete` - Network deletion
- `compute.networks.patch` - Network patch

### Subnetworks (4 events)
- `compute.subnetworks.insert` - Subnetwork creation
- `compute.subnetworks.delete` - Subnetwork deletion
- `compute.subnetworks.patch` - Subnetwork patch
- `compute.subnetworks.setPrivateIpGoogleAccess` - Private IP access change

## Example Configuration

```yaml
drift_rules:
  - name: "GCP Firewall Rule Modification"
    resource_types:
      - "google_compute_firewall"
    watched_attributes:
      - "allowed"
      - "source_ranges"
    severity: "critical"
```
