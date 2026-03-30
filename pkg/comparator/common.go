// Package comparator provides shared utilities for comparing Terraform state with actual cloud provider state.
package comparator

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// GetNestedValue retrieves a value from a nested map using dot notation (e.g., "vpc_config.subnet_ids").
// This function is identical across all cloud providers and handles nested structure navigation.
func GetNestedValue(data map[string]interface{}, path string) interface{} {
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

		// Navigate deeper
		if nextMap, ok := value.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return nil
		}
	}

	return nil
}

// ValuesEqual compares two values for equality, handling different types.
// This is the core comparison logic used by all cloud providers.
func ValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Convert to strings for comparison if types differ
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Handle boolean comparisons
	if aBool, ok := a.(bool); ok {
		if bBool, ok := b.(bool); ok {
			return aBool == bBool
		}
		// Try to parse string as boolean
		if bStr == "true" {
			return aBool
		}
		if bStr == "false" {
			return !aBool
		}
	}

	// Handle numeric comparisons
	if reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind() {
		return reflect.DeepEqual(a, b)
	}

	// Fallback to string comparison
	return aStr == bStr
}

// ValuesEqualCaseInsensitive compares two values for equality, treating strings case-insensitively.
// This is used by Azure where resource names and locations are case-insensitive.
func ValuesEqualCaseInsensitive(a, b interface{}) bool {
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

	// String-to-string: case-insensitive
	if _, aIsStr := a.(string); aIsStr {
		if _, bIsStr := b.(string); bIsStr {
			return strings.EqualFold(aStr, bStr)
		}
	}

	// Handle numeric/other comparisons with same type
	if reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind() {
		return reflect.DeepEqual(a, b)
	}

	// Fallback to case-insensitive string comparison
	return strings.EqualFold(aStr, bStr)
}

// ExtractMapStringValues extracts a map[string]string from an interface{} that is expected to be map[string]interface{}.
// This is used for extracting tags, labels, and similar string-key-value maps from attributes.
func ExtractMapStringValues(data interface{}) map[string]string {
	result := make(map[string]string)

	if mapData, ok := data.(map[string]interface{}); ok {
		for k, v := range mapData {
			if strVal, ok := v.(string); ok {
				result[k] = strVal
			}
		}
	}

	return result
}

// FilterManagedLabels removes cloud provider-managed labels/tags from the given map.
// It filters based on the provided list of ignored prefixes.
func FilterManagedLabels(labels map[string]string, ignoredPrefixes []string) map[string]string {
	filtered := make(map[string]string)

	for k, v := range labels {
		ignored := false
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(k, prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			filtered[k] = v
		}
	}

	return filtered
}

// FilterManagedLabelsCaseInsensitive removes cloud provider-managed labels/tags from the given map,
// treating the key comparison as case-insensitive. This is used by Azure.
func FilterManagedLabelsCaseInsensitive(labels map[string]string, ignoredPrefixes []string) map[string]string {
	filtered := make(map[string]string)

	for k, v := range labels {
		ignored := false
		lowerKey := strings.ToLower(k)
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(lowerKey, prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			filtered[k] = v
		}
	}

	return filtered
}

// ComparisonConfig holds provider-specific callbacks for the generic comparison orchestrator.
// Each callback allows providers to implement their own business logic while reusing the orchestration.
type ComparisonConfig struct {
	// ExtractTFID extracts the resource ID from a Terraform resource
	ExtractTFID func(tfResource interface{}) string

	// ExtractCloudID extracts the resource ID from a cloud resource
	ExtractCloudID func(cloudResource interface{}) string

	// CompareAttributes compares attributes between TF and cloud resources
	// Returns a slice of FieldDiff for any differences found
	CompareAttributes func(tfResource, cloudResource interface{}) []types.FieldDiff

	// BuildUnmanaged creates a ResourceDiff for a cloud resource not in Terraform
	BuildUnmanaged func(cloudResource interface{}) *types.ResourceDiff

	// BuildMissing creates a TerraformResource for a TF resource not in cloud
	BuildMissing func(tfResource interface{}) *types.TerraformResource

	// FindMatchingCloud is optional: finds a cloud resource matching the given TF resource
	// when exact ID match fails. Used by providers like GCP that support multiple ID formats.
	// If not provided, only exact ID matching is used.
	FindMatchingCloud func(tfResource interface{}, cloudResourceMap map[string]interface{}) interface{}
}

// CompareResources is the generic comparison orchestrator used by all cloud providers.
// It unifies the comparison logic while allowing provider-specific callbacks for details.
func CompareResources(config *ComparisonConfig, tfResources, cloudResources []interface{}) *types.DriftResult {
	result := &types.DriftResult{
		UnmanagedResources: []*types.DiscoveredResource{},
		MissingResources:   []*types.TerraformResource{},
		ModifiedResources:  []*types.ResourceDiff{},
	}

	// Build maps for efficient lookup
	tfResourceMap := make(map[string]interface{})
	for _, tfRes := range tfResources {
		resourceID := config.ExtractTFID(tfRes)
		if resourceID != "" {
			tfResourceMap[resourceID] = tfRes
		}
	}

	cloudResourceMap := make(map[string]interface{})
	for _, cloudRes := range cloudResources {
		cloudResourceMap[config.ExtractCloudID(cloudRes)] = cloudRes
	}

	log.Debugf("Comparing state: %d Terraform resources vs %d cloud resources", len(tfResourceMap), len(cloudResourceMap))

	// Track matched resources to avoid duplicate processing
	matchedCloudResources := make(map[string]bool)

	// Find missing resources (in Terraform but not in cloud) and modified resources
	for tfID, tfRes := range tfResourceMap {
		cloudRes, exists := cloudResourceMap[tfID]

		// If exact match not found and provider has fallback matching logic, try that
		if !exists && config.FindMatchingCloud != nil {
			cloudRes = config.FindMatchingCloud(tfRes, cloudResourceMap)
			if cloudRes != nil {
				// Mark this cloud resource as matched to avoid "unmanaged" classification
				matchedCloudResources[config.ExtractCloudID(cloudRes)] = true
			}
		}

		if cloudRes == nil {
			log.Debugf("Found missing resource: %s", tfID)
			result.MissingResources = append(result.MissingResources, config.BuildMissing(tfRes))
		} else {
			// Resource exists in both - check for modifications
			differences := config.CompareAttributes(tfRes, cloudRes)
			if len(differences) > 0 {
				log.Debugf("Found modified resource: %s (%d differences)", tfID, len(differences))
				// Build ResourceDiff from the differences - use the BuildUnmanaged callback to get base structure
				diff := config.BuildUnmanaged(cloudRes)
				if diff != nil {
					diff.Differences = differences
					result.ModifiedResources = append(result.ModifiedResources, diff)
				}
			}
		}
	}

	// Find unmanaged resources (in cloud but not in Terraform)
	for cloudID, cloudRes := range cloudResourceMap {
		if _, exactMatch := tfResourceMap[cloudID]; !exactMatch && !matchedCloudResources[cloudID] {
			log.Debugf("Found unmanaged resource: %s", cloudID)
			// Build unmanaged resource using provider callback
			unmanagedDiff := config.BuildUnmanaged(cloudRes)
			result.UnmanagedResources = append(result.UnmanagedResources, &types.DiscoveredResource{
				ID:         unmanagedDiff.ResourceID,
				Type:       unmanagedDiff.ResourceType,
				Provider:   unmanagedDiff.Provider,
				Attributes: unmanagedDiff.ActualState,
			})
		}
	}

	log.Debugf("Drift detection complete: %d unmanaged, %d missing, %d modified",
		len(result.UnmanagedResources), len(result.MissingResources), len(result.ModifiedResources))

	return result
}
