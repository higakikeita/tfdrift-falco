package gcp

import (
	"reflect"

	"github.com/keitahigaki/tfdrift-falco/pkg/comparator"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// CompareStateWithActual compares Terraform state with actual GCP resources
// and returns the differences (unmanaged, missing, and modified resources).
func CompareStateWithActual(tfResources []*types.TerraformResource, gcpResources []*types.DiscoveredResource) *types.DriftResult {
	config := &comparator.ComparisonConfig{
		ExtractTFID: extractTFResourceID,
		ExtractCloudID: func(cloudResource interface{}) string {
			res, ok := cloudResource.(*types.DiscoveredResource)
			if !ok {
				return ""
			}
			return res.ID
		},
		CompareAttributes: func(tfResource, cloudResource interface{}) []types.FieldDiff {
			tfRes, ok := tfResource.(*types.TerraformResource)
			if !ok {
				return nil
			}
			cloudRes, ok := cloudResource.(*types.DiscoveredResource)
			if !ok {
				return nil
			}
			diffs := compareResourceAttributes(tfRes, cloudRes)
			if diffs == nil {
				return nil
			}
			result := make([]types.FieldDiff, len(diffs))
			for i, d := range diffs {
				result[i] = *d
			}
			return result
		},
		BuildUnmanaged: func(cloudResource interface{}) *types.ResourceDiff {
			gcpRes, ok := cloudResource.(*types.DiscoveredResource)
			if !ok {
				return &types.ResourceDiff{Provider: "gcp"}
			}
			return &types.ResourceDiff{
				ResourceID:   gcpRes.ID,
				ResourceType: gcpRes.Type,
				Provider:     "gcp",
				ActualState:  gcpRes.Attributes,
				Differences:  []types.FieldDiff{},
			}
		},
		BuildMissing: func(tfResource interface{}) *types.TerraformResource {
			tfRes, ok := tfResource.(*types.TerraformResource)
			if !ok {
				return &types.TerraformResource{Provider: "gcp"}
			}
			// Extract ID from attributes if available
			resourceID := extractTFResourceID(tfRes)
			return &types.TerraformResource{
				Type:       tfRes.Type,
				Name:       tfRes.Name,
				ID:         resourceID,
				Provider:   "gcp",
				Attributes: tfRes.Attributes,
			}
		},
		// GCP supports matching by name when IDs differ
		FindMatchingCloud: func(tfResource interface{}, cloudResourceMap map[string]interface{}) interface{} {
			tfRes, ok := tfResource.(*types.TerraformResource)
			if !ok {
				return nil
			}
			return findByNameInterface(tfRes, cloudResourceMap)
		},
	}

	result := comparator.CompareResources(config, convertTFResources(tfResources), convertCloudResources(gcpResources))
	result.Provider = "gcp"
	return result
}

// convertTFResources converts []*types.TerraformResource to []interface{}
func convertTFResources(resources []*types.TerraformResource) []interface{} {
	result := make([]interface{}, len(resources))
	for i, r := range resources {
		result[i] = r
	}
	return result
}

// convertCloudResources converts []*types.DiscoveredResource to []interface{}
func convertCloudResources(resources []*types.DiscoveredResource) []interface{} {
	result := make([]interface{}, len(resources))
	for i, r := range resources {
		result[i] = r
	}
	return result
}

// extractTFResourceID extracts the GCP resource ID from Terraform resource attributes.
// GCP resources in Terraform state use various ID formats.
func extractTFResourceID(resource interface{}) string {
	tfRes, ok := resource.(*types.TerraformResource)
	if !ok {
		return ""
	}
	idFields := []string{
		"id", "self_link", "name",
	}

	for _, field := range idFields {
		if id, ok := tfRes.Attributes[field].(string); ok && id != "" {
			return id
		}
	}

	return ""
}

// compareResourceAttributes compares attributes between Terraform state and actual GCP resource.
func compareResourceAttributes(tfRes *types.TerraformResource, gcpRes *types.DiscoveredResource) []*types.FieldDiff {
	if tfRes.Type != gcpRes.Type {
		return nil
	}

	fieldsToCompare := getComparableFields(tfRes.Type)
	differences := []*types.FieldDiff{}

	for _, field := range fieldsToCompare {
		tfValue := comparator.GetNestedValue(tfRes.Attributes, field)
		gcpValue := comparator.GetNestedValue(gcpRes.Attributes, field)

		if !comparator.ValuesEqual(tfValue, gcpValue) {
			differences = append(differences, &types.FieldDiff{
				Field:          field,
				TerraformValue: tfValue,
				ActualValue:    gcpValue,
			})
		}
	}

	// Compare labels
	if !labelsEqual(tfRes.Attributes, gcpRes.Labels) {
		differences = append(differences, &types.FieldDiff{
			Field:          "labels",
			TerraformValue: getTerraformLabels(tfRes.Attributes),
			ActualValue:    gcpRes.Labels,
		})
	}

	return differences
}

// getComparableFields returns the list of fields to compare for a given GCP resource type.
func getComparableFields(resourceType string) []string {
	switch resourceType {
	case "google_compute_network":
		return []string{"auto_create_subnetworks", "routing_mode", "description"}
	case "google_compute_subnetwork":
		return []string{"ip_cidr_range", "network", "private_ip_google_access"}
	case "google_compute_firewall":
		return []string{"network", "direction", "priority", "disabled"}
	case "google_compute_instance":
		return []string{"machine_type", "zone", "status"}
	case "google_storage_bucket":
		return []string{"location", "storage_class", "versioning"}
	case "google_sql_database_instance":
		return []string{"database_version", "region", "tier", "availability_type"}
	case "google_container_cluster":
		return []string{"location", "network", "subnetwork", "current_master_version"}
	case "google_cloud_run_v2_service":
		return []string{"location", "ingress"}
	default:
		return []string{}
	}
}

// Internal wrapper functions for test compatibility.
// These delegate to the shared comparator package functions.

// getNestedValue is a wrapper around comparator.GetNestedValue for test compatibility.
func getNestedValue(data map[string]interface{}, path string) interface{} {
	return comparator.GetNestedValue(data, path)
}

// valuesEqual is a wrapper around comparator.ValuesEqual for test compatibility.
func valuesEqual(a, b interface{}) bool {
	return comparator.ValuesEqual(a, b)
}

// labelsEqual compares labels between Terraform state and GCP (provider-specific logic)
func labelsEqual(tfAttrs map[string]interface{}, gcpLabels map[string]string) bool {
	tfLabels := getTerraformLabels(tfAttrs)

	// Ignore GCP-managed labels
	ignoredPrefixes := []string{"goog-", "gke-"}

	filteredGCPLabels := comparator.FilterManagedLabels(gcpLabels, ignoredPrefixes)

	return reflect.DeepEqual(tfLabels, filteredGCPLabels)
}

// getTerraformLabels extracts labels from Terraform resource attributes.
func getTerraformLabels(attrs map[string]interface{}) map[string]string {
	labels := make(map[string]string)

	if labelsVal, ok := attrs["labels"].(map[string]interface{}); ok {
		for k, v := range labelsVal {
			if strVal, ok := v.(string); ok {
				labels[k] = strVal
			}
		}
		return labels
	}

	return labels
}

// matchesByName checks if a GCP resource matches any Terraform resource by name.
// This is used for resources where ID format differs between TF and GCP.
func matchesByName(gcpRes *types.DiscoveredResource, tfMap map[string]interface{}) bool {
	for _, tfRes := range tfMap {
		tfr, ok := tfRes.(*types.TerraformResource)
		if !ok {
			continue
		}
		if tfr.Type == gcpRes.Type {
			if name, ok := tfr.Attributes["name"].(string); ok && name == gcpRes.Name {
				return true
			}
		}
	}
	return false
}

// findByName finds a GCP resource matching the Terraform resource by name and type.
// This is used for resources where ID format differs between TF and GCP.
func findByName(tfRes *types.TerraformResource, gcpMap map[string]*types.DiscoveredResource) *types.DiscoveredResource {
	tfName, _ := tfRes.Attributes["name"].(string)
	if tfName == "" {
		return nil
	}

	for _, gcpRes := range gcpMap {
		if gcpRes.Type == tfRes.Type && gcpRes.Name == tfName {
			return gcpRes
		}
	}
	return nil
}

// findByNameInterface is a wrapper around findByName for the generic interface{} API
func findByNameInterface(tfRes *types.TerraformResource, gcpMap map[string]interface{}) interface{} {
	tfName, _ := tfRes.Attributes["name"].(string)
	if tfName == "" {
		return nil
	}

	for _, gcpResInterface := range gcpMap {
		gcpRes, ok := gcpResInterface.(*types.DiscoveredResource)
		if !ok {
			continue
		}
		if gcpRes.Type == tfRes.Type && gcpRes.Name == tfName {
			return gcpRes
		}
	}
	return nil
}
