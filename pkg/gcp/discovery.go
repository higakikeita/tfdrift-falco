package gcp

import (
	"context"
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
	run "google.golang.org/api/run/v2"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

// DiscoveryClient handles GCP resource discovery across a project.
type DiscoveryClient struct {
	projectID string
	regions   []string // GCP regions to scan

	computeService   *compute.Service
	containerService *container.Service
	sqlService       *sqladmin.Service
	storageClient    *storage.Client
	runService       *run.Service
}

// DiscoveredResource represents a resource found in GCP.
type DiscoveredResource struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // Terraform resource type
	Name       string                 `json:"name"`
	Region     string                 `json:"region"`
	SelfLink   string                 `json:"self_link,omitempty"`
	Attributes map[string]interface{} `json:"attributes"`
	Labels     map[string]string      `json:"labels,omitempty"`
}

// DriftResult represents the difference between Terraform and actual GCP state.
type DriftResult struct {
	UnmanagedResources []*DiscoveredResource `json:"unmanaged_resources"`
	MissingResources   []*TerraformResource  `json:"missing_resources"`
	ModifiedResources  []*ResourceDiff       `json:"modified_resources"`
}

// TerraformResource is a minimal representation for GCP drift results.
type TerraformResource struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ResourceDiff represents differences in a single resource.
type ResourceDiff struct {
	ResourceID     string                 `json:"resource_id"`
	ResourceType   string                 `json:"resource_type"`
	TerraformState map[string]interface{} `json:"terraform_state"`
	ActualState    map[string]interface{} `json:"actual_state"`
	Differences    []FieldDiff            `json:"differences"`
}

// FieldDiff represents a difference in a specific field.
type FieldDiff struct {
	Field          string      `json:"field"`
	TerraformValue interface{} `json:"terraform_value"`
	ActualValue    interface{} `json:"actual_value"`
}

// NewDiscoveryClient creates a new GCP discovery client.
// projectID is the GCP project to discover resources in.
// regions specifies which regions to scan (if empty, discovers across all zones in the project).
func NewDiscoveryClient(ctx context.Context, projectID string, regions []string, opts ...option.ClientOption) (*DiscoveryClient, error) {
	computeSvc, err := compute.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Compute service: %w", err)
	}

	containerSvc, err := container.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Container service: %w", err)
	}

	sqlSvc, err := sqladmin.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL Admin service: %w", err)
	}

	storageC, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Storage client: %w", err)
	}

	runSvc, err := run.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run service: %w", err)
	}

	return &DiscoveryClient{
		projectID:        projectID,
		regions:          regions,
		computeService:   computeSvc,
		containerService: containerSvc,
		sqlService:       sqlSvc,
		storageClient:    storageC,
		runService:       runSvc,
	}, nil
}

// DiscoverAll discovers all supported GCP resources in the project.
func (d *DiscoveryClient) DiscoverAll(ctx context.Context) ([]*DiscoveredResource, error) {
	log.Infof("Starting GCP resource discovery in project %s", d.projectID)

	var allResources []*DiscoveredResource

	// Discover VPC Networks (global)
	networks, err := d.discoverNetworks(ctx)
	if err != nil {
		log.Warnf("Failed to discover VPC Networks: %v", err)
	} else {
		allResources = append(allResources, networks...)
		log.Infof("Discovered %d VPC Networks", len(networks))
	}

	// Discover Subnetworks (regional)
	subnets, err := d.discoverSubnetworks(ctx)
	if err != nil {
		log.Warnf("Failed to discover Subnetworks: %v", err)
	} else {
		allResources = append(allResources, subnets...)
		log.Infof("Discovered %d Subnetworks", len(subnets))
	}

	// Discover Firewall Rules (global)
	firewalls, err := d.discoverFirewalls(ctx)
	if err != nil {
		log.Warnf("Failed to discover Firewalls: %v", err)
	} else {
		allResources = append(allResources, firewalls...)
		log.Infof("Discovered %d Firewalls", len(firewalls))
	}

	// Discover Compute Instances (zonal)
	instances, err := d.discoverInstances(ctx)
	if err != nil {
		log.Warnf("Failed to discover Compute Instances: %v", err)
	} else {
		allResources = append(allResources, instances...)
		log.Infof("Discovered %d Compute Instances", len(instances))
	}

	// Discover GCS Buckets (global)
	buckets, err := d.discoverBuckets(ctx)
	if err != nil {
		log.Warnf("Failed to discover GCS Buckets: %v", err)
	} else {
		allResources = append(allResources, buckets...)
		log.Infof("Discovered %d GCS Buckets", len(buckets))
	}

	// Discover Cloud SQL Instances
	sqlInstances, err := d.discoverSQLInstances(ctx)
	if err != nil {
		log.Warnf("Failed to discover Cloud SQL Instances: %v", err)
	} else {
		allResources = append(allResources, sqlInstances...)
		log.Infof("Discovered %d Cloud SQL Instances", len(sqlInstances))
	}

	// Discover GKE Clusters
	clusters, err := d.discoverGKEClusters(ctx)
	if err != nil {
		log.Warnf("Failed to discover GKE Clusters: %v", err)
	} else {
		allResources = append(allResources, clusters...)
		log.Infof("Discovered %d GKE Clusters", len(clusters))
	}

	// Discover Cloud Run Services
	runServices, err := d.discoverCloudRunServices(ctx)
	if err != nil {
		log.Warnf("Failed to discover Cloud Run Services: %v", err)
	} else {
		allResources = append(allResources, runServices...)
		log.Infof("Discovered %d Cloud Run Services", len(runServices))
	}

	log.Infof("GCP discovery completed: %d total resources discovered", len(allResources))
	return allResources, nil
}

// discoverNetworks discovers all VPC Networks in the project.
func (d *DiscoveryClient) discoverNetworks(ctx context.Context) ([]*DiscoveredResource, error) {
	networkList, err := d.computeService.Networks.List(d.projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	var resources []*DiscoveredResource
	for _, n := range networkList.Items {
		resources = append(resources, &DiscoveredResource{
			ID:       fmt.Sprintf("projects/%s/global/networks/%s", d.projectID, n.Name),
			Type:     "google_compute_network",
			Name:     n.Name,
			Region:   "global",
			SelfLink: n.SelfLink,
			Attributes: map[string]interface{}{
				"name":                    n.Name,
				"auto_create_subnetworks": n.AutoCreateSubnetworks,
				"routing_mode":            n.RoutingConfig.RoutingMode,
				"description":             n.Description,
			},
		})
	}
	return resources, nil
}

// discoverSubnetworks discovers all subnetworks, optionally filtered by region.
func (d *DiscoveryClient) discoverSubnetworks(ctx context.Context) ([]*DiscoveredResource, error) {
	var resources []*DiscoveredResource

	if len(d.regions) > 0 {
		for _, region := range d.regions {
			subs, err := d.computeService.Subnetworks.List(d.projectID, region).Context(ctx).Do()
			if err != nil {
				log.Warnf("Failed to list subnetworks in region %s: %v", region, err)
				continue
			}
			for _, s := range subs.Items {
				resources = append(resources, subnetworkToDiscovered(d.projectID, s))
			}
		}
	} else {
		// Use aggregatedList to get all subnetworks across all regions
		aggList, err := d.computeService.Subnetworks.AggregatedList(d.projectID).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to aggregated list subnetworks: %w", err)
		}
		for _, scopedList := range aggList.Items {
			for _, s := range scopedList.Subnetworks {
				resources = append(resources, subnetworkToDiscovered(d.projectID, s))
			}
		}
	}

	return resources, nil
}

func subnetworkToDiscovered(projectID string, s *compute.Subnetwork) *DiscoveredResource {
	region := extractRegionFromURL(s.Region)
	return &DiscoveredResource{
		ID:       fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", projectID, region, s.Name),
		Type:     "google_compute_subnetwork",
		Name:     s.Name,
		Region:   region,
		SelfLink: s.SelfLink,
		Attributes: map[string]interface{}{
			"name":                       s.Name,
			"network":                    s.Network,
			"ip_cidr_range":              s.IpCidrRange,
			"region":                     region,
			"private_ip_google_access":   s.PrivateIpGoogleAccess,
			"purpose":                    s.Purpose,
		},
	}
}

// discoverFirewalls discovers all firewall rules in the project.
func (d *DiscoveryClient) discoverFirewalls(ctx context.Context) ([]*DiscoveredResource, error) {
	fwList, err := d.computeService.Firewalls.List(d.projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list firewalls: %w", err)
	}

	var resources []*DiscoveredResource
	for _, fw := range fwList.Items {
		attrs := map[string]interface{}{
			"name":        fw.Name,
			"network":     fw.Network,
			"direction":   fw.Direction,
			"priority":    fw.Priority,
			"description": fw.Description,
			"disabled":    fw.Disabled,
		}
		if len(fw.SourceRanges) > 0 {
			attrs["source_ranges"] = fw.SourceRanges
		}
		if len(fw.DestinationRanges) > 0 {
			attrs["destination_ranges"] = fw.DestinationRanges
		}
		if len(fw.SourceTags) > 0 {
			attrs["source_tags"] = fw.SourceTags
		}
		if len(fw.TargetTags) > 0 {
			attrs["target_tags"] = fw.TargetTags
		}

		resources = append(resources, &DiscoveredResource{
			ID:         fmt.Sprintf("projects/%s/global/firewalls/%s", d.projectID, fw.Name),
			Type:       "google_compute_firewall",
			Name:       fw.Name,
			Region:     "global",
			SelfLink:   fw.SelfLink,
			Attributes: attrs,
		})
	}
	return resources, nil
}

// discoverInstances discovers all Compute Engine instances.
func (d *DiscoveryClient) discoverInstances(ctx context.Context) ([]*DiscoveredResource, error) {
	var resources []*DiscoveredResource

	aggList, err := d.computeService.Instances.AggregatedList(d.projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to aggregated list instances: %w", err)
	}

	for _, scopedList := range aggList.Items {
		for _, inst := range scopedList.Instances {
			zone := extractZoneFromURL(inst.Zone)
			region := zoneToRegion(zone)

			// Skip if region filter is set and doesn't match
			if len(d.regions) > 0 && !containsString(d.regions, region) {
				continue
			}

			attrs := map[string]interface{}{
				"name":         inst.Name,
				"machine_type": extractLastSegment(inst.MachineType),
				"zone":         zone,
				"status":       inst.Status,
				"description":  inst.Description,
			}

			// Extract network interfaces
			if len(inst.NetworkInterfaces) > 0 {
				ni := inst.NetworkInterfaces[0]
				attrs["network"] = ni.Network
				attrs["subnetwork"] = ni.Subnetwork
				attrs["network_ip"] = ni.NetworkIP
			}

			resources = append(resources, &DiscoveredResource{
				ID:         fmt.Sprintf("projects/%s/zones/%s/instances/%s", d.projectID, zone, inst.Name),
				Type:       "google_compute_instance",
				Name:       inst.Name,
				Region:     region,
				SelfLink:   inst.SelfLink,
				Attributes: attrs,
				Labels:     inst.Labels,
			})
		}
	}

	return resources, nil
}

// discoverBuckets discovers all GCS Buckets in the project.
func (d *DiscoveryClient) discoverBuckets(ctx context.Context) ([]*DiscoveredResource, error) {
	var resources []*DiscoveredResource

	it := d.storageClient.Buckets(ctx, d.projectID)
	for {
		bucket, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate buckets: %w", err)
		}

		resources = append(resources, &DiscoveredResource{
			ID:   bucket.Name,
			Type: "google_storage_bucket",
			Name: bucket.Name,
			Region: func() string {
				if bucket.Location != "" {
					return strings.ToLower(bucket.Location)
				}
				return "global"
			}(),
			Attributes: map[string]interface{}{
				"name":          bucket.Name,
				"location":      strings.ToLower(bucket.Location),
				"storage_class": bucket.StorageClass,
				"versioning":    bucket.VersioningEnabled,
			},
			Labels: bucket.Labels,
		})
	}

	return resources, nil
}

// discoverSQLInstances discovers all Cloud SQL instances.
func (d *DiscoveryClient) discoverSQLInstances(ctx context.Context) ([]*DiscoveredResource, error) {
	sqlList, err := d.sqlService.Instances.List(d.projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list SQL instances: %w", err)
	}

	var resources []*DiscoveredResource
	for _, inst := range sqlList.Items {
		region := inst.Region
		if region == "" {
			region = inst.GceZone
		}

		attrs := map[string]interface{}{
			"name":             inst.Name,
			"database_version": inst.DatabaseVersion,
			"region":           region,
			"state":            inst.State,
			"connection_name":  inst.ConnectionName,
		}

		if inst.Settings != nil {
			attrs["tier"] = inst.Settings.Tier
			attrs["availability_type"] = inst.Settings.AvailabilityType
			attrs["disk_size"] = inst.Settings.DataDiskSizeGb
			attrs["disk_type"] = inst.Settings.DataDiskType
		}

		resources = append(resources, &DiscoveredResource{
			ID:         fmt.Sprintf("projects/%s/instances/%s", d.projectID, inst.Name),
			Type:       "google_sql_database_instance",
			Name:       inst.Name,
			Region:     region,
			SelfLink:   inst.SelfLink,
			Attributes: attrs,
			Labels:     inst.Settings.UserLabels,
		})
	}
	return resources, nil
}

// discoverGKEClusters discovers all GKE clusters.
func (d *DiscoveryClient) discoverGKEClusters(ctx context.Context) ([]*DiscoveredResource, error) {
	parent := fmt.Sprintf("projects/%s/locations/-", d.projectID)
	resp, err := d.containerService.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list GKE clusters: %w", err)
	}

	var resources []*DiscoveredResource
	for _, cluster := range resp.Clusters {
		region := cluster.Location

		// Skip if region filter is set and doesn't match
		if len(d.regions) > 0 && !containsString(d.regions, region) && !containsString(d.regions, zoneToRegion(region)) {
			continue
		}

		attrs := map[string]interface{}{
			"name":                  cluster.Name,
			"location":              cluster.Location,
			"network":               cluster.Network,
			"subnetwork":            cluster.Subnetwork,
			"cluster_ipv4_cidr":     cluster.ClusterIpv4Cidr,
			"services_ipv4_cidr":    cluster.ServicesIpv4Cidr,
			"current_master_version": cluster.CurrentMasterVersion,
			"current_node_version":  cluster.CurrentNodeVersion,
			"status":                cluster.Status,
			"initial_node_count":    cluster.InitialNodeCount,
		}

		resources = append(resources, &DiscoveredResource{
			ID:         fmt.Sprintf("projects/%s/locations/%s/clusters/%s", d.projectID, cluster.Location, cluster.Name),
			Type:       "google_container_cluster",
			Name:       cluster.Name,
			Region:     region,
			SelfLink:   cluster.SelfLink,
			Attributes: attrs,
			Labels:     cluster.ResourceLabels,
		})
	}
	return resources, nil
}

// discoverCloudRunServices discovers all Cloud Run services.
func (d *DiscoveryClient) discoverCloudRunServices(ctx context.Context) ([]*DiscoveredResource, error) {
	var resources []*DiscoveredResource

	parent := fmt.Sprintf("projects/%s/locations/-", d.projectID)
	resp, err := d.runService.Projects.Locations.Services.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud Run services: %w", err)
	}

	for _, svc := range resp.Services {
		// Extract region from name: projects/PROJECT/locations/REGION/services/NAME
		parts := strings.Split(svc.Name, "/")
		region := ""
		serviceName := ""
		if len(parts) >= 6 {
			region = parts[3]
			serviceName = parts[5]
		}

		// Skip if region filter is set and doesn't match
		if len(d.regions) > 0 && region != "" && !containsString(d.regions, region) {
			continue
		}

		attrs := map[string]interface{}{
			"name":     serviceName,
			"location": region,
			"uri":      svc.Uri,
			"ingress":  svc.Ingress,
		}
		if svc.Template != nil {
			if svc.Template.ServiceAccount != "" {
				attrs["service_account"] = svc.Template.ServiceAccount
			}
			if svc.Template.MaxInstanceRequestConcurrency > 0 {
				attrs["max_instance_request_concurrency"] = svc.Template.MaxInstanceRequestConcurrency
			}
		}

		resources = append(resources, &DiscoveredResource{
			ID:         svc.Name,
			Type:       "google_cloud_run_v2_service",
			Name:       serviceName,
			Region:     region,
			Attributes: attrs,
			Labels:     svc.Labels,
		})
	}
	return resources, nil
}

// --- Helper functions ---

// extractRegionFromURL extracts region name from a GCP resource URL.
// e.g., "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1" -> "us-central1"
func extractRegionFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, p := range parts {
		if p == "regions" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return url
}

// extractZoneFromURL extracts zone name from a GCP resource URL.
func extractZoneFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, p := range parts {
		if p == "zones" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return url
}

// extractLastSegment extracts the last path segment from a URL.
// e.g., ".../machineTypes/e2-medium" -> "e2-medium"
func extractLastSegment(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}

// zoneToRegion converts a GCP zone to its parent region.
// e.g., "us-central1-a" -> "us-central1"
func zoneToRegion(zone string) string {
	parts := strings.Split(zone, "-")
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-1], "-")
	}
	return zone
}

// containsString checks if a string slice contains a given string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
