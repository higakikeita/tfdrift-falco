package gcp

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CompareStateWithActual compares Terraform state with actual GCP resources
// and returns the differences (unmanaged, missing, and modified resources).
func CompareStateWithActual(tfResources []*TerraformResource, gcpResources []*DiscoveredResource) *DriftResult {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{},
		MissingResources:   []*TerraformResource{},
		ModifiedResources:  []*ResourceDiff{},
	}

	// Build maps for efficient lookup
	tfResourceMap := make(map[string]*TerraformResource)
	for _, tfRes := range tfResources {
		resourceID := extractTFResourceID(tfRes)
		if resourceID != "" {
			tfResourceMap[resourceID] = tfRes
		}
	}

	gcpResourceMap := make(map[string]*DiscoveredResource)
	for _, gcpRes := range gcpResources {
		gcpResourceMap[gcpRes.ID] = gcpRes
	}

	log.Infof("Comparing state: %d Terraform resources vs %d GCP resources", len(tfResourceMap), len(gcpResourceMap))

	// Find unmanaged resources (in GCP but not in Terraform)
	for gcpID, gcpRes := range gcpResourceMap {
		if _, exists := tfResourceMap[gcpID]; !exists {
			// Also try matching by name for resources where ID format differs
			if !matchesByName(gcpRes, tfResourceMap) {
				log.Infof("Found unmanaged resource: %s (%s)", gcpID, gcpRes.Type)
				result.UnmanagedResources = append(result.UnmanagedResources, gcpRes)
			}
		}
	}

	// Find missing resources (in Terraform but not in GCP) and modified resources
	for tfID, tfRes := range tfResourceMap {
		gcpRes, exists := gcpResourceMap[tfID]
		if !exists {
			// Try matching by name
			gcpRes = findByName(tfRes, gcpResourceMap)
		}

		if gcpRes == nil {
			log.Infof("Found missing resource: %s (%s)", tfID, tfRes.Type)
			result.MissingResources = append(result.MissingResources, tfRes)
		} else {
			// Resource exists in both - check for modifications
			diff := compareResourceAttributes(tfRes, gcpRes)
			if diff != nil && len(diff.Differences) > 0 {
				log.Infof("Found modified resource: %s (%d differences)", tfID, len(diff.Differences))
				result.ModifiedResources = append(result.ModifiedResources, diff)
			}
		}
	}

	log.Infof("GCP drift detection complete: %d unmanaged, %d missing, %d modified",
		len(result.UnmanagedResources), len(result.MissingResources), len(result.ModifiedResources))

	return result
}

// extractTFResourceID extracts the GCP resource ID from Terraform resource attributes.
// GCP resources in Terraform state use various ID formats.
func extractTFResourceID(tfRes *TerraformResource) string {
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

// matchesByName checks if a GCP resource matches any Terraform resource by name.
func matchesByName(gcpRes *DiscoveredResource, tfMap map[string]*TerraformResource) bool {
	for _, tfRes := range tfMap {
		if tfRes.Type == gcpRes.Type {
			if name, ok := tfRes.Attributes["name"].(string); ok && name == gcpRes.Name {
				return true
			}
		}
	}
	return false
}

// findByName finds a GCP resource matching the Terraform resource by name and type.
func findByName(tfRes *TerraformResource, gcpMap map[string]*DiscoveredResource) *DiscoveredResource {
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

// compareResourceAttributes compares attributes between Terraform state and actual GCP resource.
func compareResourceAttributes(tfRes *TerraformResource, gcpRes *DiscoveredResource) *ResourceDiff {
	if tfRes.Type != gcpRes.Type {
		return nil
	}

	diff := &ResourceDiff{
		ResourceID:     gcpRes.ID,
		ResourceType:   tfRes.Type,
		TerraformState: tfRes.Attributes,
		ActualState:    gcpRes.Attributes,
		Differences:    []FieldDiff{},
	}

	fieldsToCompare := getComparableFields(tfRes.Type)

	for _, field := range fieldsToCompare {
		tfValue := getNestedValue(tfRes.Attributes, field)
		gcpValue := getNestedValue(gcpRes.Attributes, field)

		if !valuesEqual(tfValue, gcpValue) {
			diff.Differences = append(diff.Differences, FieldDiff{
				Field:          field,
				TerraformValue: tfValue,
				ActualValue:    gcpValue,
			})
		}
	}

	// Compare labels
	if !labelsEqual(tfRes.Attributes, gcpRes.Labels) {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:          "labels",
			TerraformValue: getTerraformLabels(tfRes.Attributes),
			ActualValue:    gcpRes.Labels,
		})
	}

	return diff
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

// getNestedValue retrieves a value from a nested map using dot notation.
func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil
		}

		if i == len(parts)-1 {
			return value
		}

		if nextMap, ok := value.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return nil
		}
	}

	return nil
}

// valuesEqual compares two values for equality, handling different types.
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try string comparison for mixed types
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Handle boolean comparisons
	if aBool, ok := a.(bool); ok {
		if bStr == "true" {
			return aBool
		}
		if bStr == "false" {
			return !aBool
		}
	}

	// Handle numeric comparisons with same type
	if reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind() {
		return reflect.DeepEqual(a, b)
	}

	// Fallback to string comparison
	return aStr == bStr
}

// labelsEqual compares labels between Terraform state and GCP.
func labelsEqual(tfAttrs map[string]interface{}, gcpLabels map[string]string) bool {
	tfLabels := getTerraformLabels(tfAttrs)

	// Ignore GCP-managed labels
	ignoredPrefixes := []string{"goog-", "gke-"}

	filteredGCPLabels := make(map[string]string)
	for k, v := range gcpLabels {
		ignored := false
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(k, prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			filteredGCPLabels[k] = v
		}
	}

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
