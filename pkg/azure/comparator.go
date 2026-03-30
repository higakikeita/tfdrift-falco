package azure

import (
	"reflect"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/comparator"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// CompareStateWithActual compares Terraform state with actual Azure resources
// and returns the differences (unmanaged, missing, and modified resources).
func CompareStateWithActual(tfResources []*types.TerraformResource, azureResources []*types.DiscoveredResource) *types.DriftResult {
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
			azRes, ok := cloudResource.(*types.DiscoveredResource)
			if !ok {
				return &types.ResourceDiff{Provider: "azure"}
			}
			return &types.ResourceDiff{
				ResourceID:    azRes.ID,
				ResourceType:  azRes.Type,
				Provider:      "azure",
				ActualState:   azRes.Attributes,
				Differences:   []types.FieldDiff{},
			}
		},
		BuildMissing: func(tfResource interface{}) *types.TerraformResource {
			tfRes, ok := tfResource.(*types.TerraformResource)
			if !ok {
				return &types.TerraformResource{Provider: "azure"}
			}
			// Extract ID from attributes if available
			resourceID := extractTFResourceID(tfRes)
			return &types.TerraformResource{
				Type:       tfRes.Type,
				Name:       tfRes.Name,
				ID:         resourceID,
				Provider:   "azure",
				Attributes: tfRes.Attributes,
			}
		},
		// Azure supports case-insensitive matching by name when IDs differ
		FindMatchingCloud: func(tfResource interface{}, cloudResourceMap map[string]interface{}) interface{} {
			tfRes, ok := tfResource.(*types.TerraformResource)
			if !ok {
				return nil
			}
			return findByNameInterface(tfRes, cloudResourceMap)
		},
	}

	result := comparator.CompareResources(config, convertTFResources(tfResources), convertCloudResources(azureResources))
	result.Provider = "azure"
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

// extractTFResourceID extracts the Azure resource ID from Terraform resource attributes.
func extractTFResourceID(resource interface{}) string {
	tfRes, ok := resource.(*types.TerraformResource)
	if !ok {
		return ""
	}
	idFields := []string{
		"id", "name",
	}

	for _, field := range idFields {
		if id, ok := tfRes.Attributes[field].(string); ok && id != "" {
			return id
		}
	}

	return ""
}

// compareResourceAttributes compares attributes between Terraform state and actual Azure resource.
func compareResourceAttributes(tfRes *types.TerraformResource, azRes *types.DiscoveredResource) []*types.FieldDiff {
	if tfRes.Type != azRes.Type {
		return nil
	}

	fieldsToCompare := getComparableFields(tfRes.Type)
	differences := []*types.FieldDiff{}

	for _, field := range fieldsToCompare {
		tfValue := comparator.GetNestedValue(tfRes.Attributes, field)
		azValue := comparator.GetNestedValue(azRes.Attributes, field)

		if !comparator.ValuesEqualCaseInsensitive(tfValue, azValue) {
			differences = append(differences, &types.FieldDiff{
				Field:          field,
				TerraformValue: tfValue,
				ActualValue:    azValue,
			})
		}
	}

	// Compare tags
	if !tagsEqual(tfRes.Attributes, azRes.Tags) {
		differences = append(differences, &types.FieldDiff{
			Field:          "tags",
			TerraformValue: getTerraformTags(tfRes.Attributes),
			ActualValue:    azRes.Tags,
		})
	}

	return differences
}

// getComparableFields returns the list of fields to compare for a given Azure resource type.
func getComparableFields(resourceType string) []string {
	switch resourceType {
	case "azurerm_virtual_machine":
		return []string{"vm_size", "location", "provisioning_state"}
	case "azurerm_virtual_machine_scale_set":
		return []string{"location", "sku_name", "sku_capacity"}
	case "azurerm_virtual_network":
		return []string{"location", "address_space"}
	case "azurerm_network_security_group":
		return []string{"location"}
	case "azurerm_storage_account":
		return []string{"location", "sku_name"}
	case "azurerm_kubernetes_cluster":
		return []string{"location", "kubernetes_version", "dns_prefix"}
	case "azurerm_mssql_server":
		return []string{"location", "version", "administrator_login"}
	case "azurerm_mssql_database":
		return []string{"collation", "max_size_gb"}
	case "azurerm_key_vault":
		return []string{"location", "sku_name"}
	case "azurerm_cosmosdb_account":
		return []string{"location", "offer_type", "consistency_level"}
	case "azurerm_lb":
		return []string{"location", "sku_name"}
	case "azurerm_public_ip":
		return []string{"location", "allocation_method", "ip_address"}
	case "azurerm_app_service":
		return []string{"location", "state"}
	case "azurerm_app_service_plan":
		return []string{"location", "sku_name"}
	case "azurerm_redis_cache":
		return []string{"location", "hostname", "port", "ssl_port"}
	case "azurerm_container_registry":
		return []string{"location", "sku_name", "admin_enabled", "login_server"}
	case "azurerm_managed_disk":
		return []string{"location", "sku_name"}
	case "azurerm_dns_zone":
		return []string{"location"}
	case "azurerm_resource_group":
		return []string{"location"}
	case "azurerm_servicebus_namespace":
		return []string{"location", "sku_name"}
	case "azurerm_eventhub_namespace":
		return []string{"location", "sku_name"}
	case "azurerm_log_analytics_workspace":
		return []string{"location", "sku_name"}
	case "azurerm_application_insights":
		return []string{"location"}
	case "azurerm_firewall":
		return []string{"location", "sku_name"}
	case "azurerm_route_table":
		return []string{"location"}
	case "azurerm_network_interface":
		return []string{"location"}
	default:
		return []string{"location"}
	}
}

// Internal wrapper functions for test compatibility.
// These delegate to the shared comparator package functions.

// getNestedValue is a wrapper around comparator.GetNestedValue for test compatibility.
func getNestedValue(data map[string]interface{}, path string) interface{} {
	return comparator.GetNestedValue(data, path)
}

// valuesEqual is a wrapper around comparator.ValuesEqualCaseInsensitive for test compatibility.
// Azure uses case-insensitive comparison.
func valuesEqual(a, b interface{}) bool {
	return comparator.ValuesEqualCaseInsensitive(a, b)
}

// tagsEqual compares tags between Terraform state and Azure (provider-specific logic)
func tagsEqual(tfAttrs map[string]interface{}, azureTags map[string]string) bool {
	tfTags := getTerraformTags(tfAttrs)

	// Ignore Azure-managed tags
	ignoredPrefixes := []string{"hidden-", "ms-resource-usage"}

	filteredAzureTags := comparator.FilterManagedLabelsCaseInsensitive(azureTags, ignoredPrefixes)

	return reflect.DeepEqual(tfTags, filteredAzureTags)
}

// getTerraformTags extracts tags from Terraform resource attributes.
func getTerraformTags(attrs map[string]interface{}) map[string]string {
	tags := make(map[string]string)

	if tagsVal, ok := attrs["tags"].(map[string]interface{}); ok {
		for k, v := range tagsVal {
			if strVal, ok := v.(string); ok {
				tags[k] = strVal
			}
		}
		return tags
	}

	return tags
}

// matchesByName checks if an Azure resource matches any Terraform resource by name.
// This is used for resources where ID format differs between TF and Azure.
func matchesByName(azRes *types.DiscoveredResource, tfMap map[string]*types.TerraformResource) bool {
	for _, tfRes := range tfMap {
		if tfRes.Type == azRes.Type {
			if name, ok := tfRes.Attributes["name"].(string); ok && strings.EqualFold(name, azRes.Name) {
				return true
			}
		}
	}
	return false
}

// findByName finds an Azure resource matching the Terraform resource by name and type.
// This is used for resources where ID format differs between TF and Azure.
func findByName(tfRes *types.TerraformResource, azureMap map[string]*types.DiscoveredResource) *types.DiscoveredResource {
	tfName, _ := tfRes.Attributes["name"].(string)
	if tfName == "" {
		return nil
	}

	for _, azRes := range azureMap {
		if azRes.Type == tfRes.Type && strings.EqualFold(azRes.Name, tfName) {
			return azRes
		}
	}
	return nil
}

// findByNameInterface is a wrapper around findByName for the generic interface{} API
func findByNameInterface(tfRes *types.TerraformResource, azureMap map[string]interface{}) interface{} {
	tfName, _ := tfRes.Attributes["name"].(string)
	if tfName == "" {
		return nil
	}

	for _, azResInterface := range azureMap {
		azRes, ok := azResInterface.(*types.DiscoveredResource)
		if !ok {
			continue
		}
		if azRes.Type == tfRes.Type && strings.EqualFold(azRes.Name, tfName) {
			return azRes
		}
	}
	return nil
}
