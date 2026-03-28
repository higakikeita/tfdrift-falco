package azure

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CompareStateWithActual compares Terraform state with actual Azure resources
// and returns the differences (unmanaged, missing, and modified resources).
func CompareStateWithActual(tfResources []*TerraformResource, azureResources []*DiscoveredResource) *DriftResult {
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

	azureResourceMap := make(map[string]*DiscoveredResource)
	for _, azRes := range azureResources {
		azureResourceMap[azRes.ID] = azRes
	}

	log.Infof("Comparing state: %d Terraform resources vs %d Azure resources", len(tfResourceMap), len(azureResourceMap))

	// Find unmanaged resources (in Azure but not in Terraform)
	for azID, azRes := range azureResourceMap {
		if _, exists := tfResourceMap[azID]; !exists {
			// Also try matching by name for resources where ID format differs
			if !matchesByName(azRes, tfResourceMap) {
				log.Infof("Found unmanaged resource: %s (%s)", azID, azRes.Type)
				result.UnmanagedResources = append(result.UnmanagedResources, azRes)
			}
		}
	}

	// Find missing resources (in Terraform but not in Azure) and modified resources
	for tfID, tfRes := range tfResourceMap {
		azRes, exists := azureResourceMap[tfID]
		if !exists {
			// Try matching by name
			azRes = findByName(tfRes, azureResourceMap)
		}

		if azRes == nil {
			log.Infof("Found missing resource: %s (%s)", tfID, tfRes.Type)
			result.MissingResources = append(result.MissingResources, tfRes)
		} else {
			// Resource exists in both - check for modifications
			diff := compareResourceAttributes(tfRes, azRes)
			if diff != nil && len(diff.Differences) > 0 {
				log.Infof("Found modified resource: %s (%d differences)", tfID, len(diff.Differences))
				result.ModifiedResources = append(result.ModifiedResources, diff)
			}
		}
	}

	log.Infof("Azure drift detection complete: %d unmanaged, %d missing, %d modified",
		len(result.UnmanagedResources), len(result.MissingResources), len(result.ModifiedResources))

	return result
}

// extractTFResourceID extracts the Azure resource ID from Terraform resource attributes.
func extractTFResourceID(tfRes *TerraformResource) string {
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

// matchesByName checks if an Azure resource matches any Terraform resource by name.
func matchesByName(azRes *DiscoveredResource, tfMap map[string]*TerraformResource) bool {
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
func findByName(tfRes *TerraformResource, azureMap map[string]*DiscoveredResource) *DiscoveredResource {
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

// compareResourceAttributes compares attributes between Terraform state and actual Azure resource.
func compareResourceAttributes(tfRes *TerraformResource, azRes *DiscoveredResource) *ResourceDiff {
	if tfRes.Type != azRes.Type {
		return nil
	}

	diff := &ResourceDiff{
		ResourceID:     azRes.ID,
		ResourceType:   tfRes.Type,
		TerraformState: tfRes.Attributes,
		ActualState:    azRes.Attributes,
		Differences:    []FieldDiff{},
	}

	fieldsToCompare := getComparableFields(tfRes.Type)

	for _, field := range fieldsToCompare {
		tfValue := getNestedValue(tfRes.Attributes, field)
		azValue := getNestedValue(azRes.Attributes, field)

		if !valuesEqual(tfValue, azValue) {
			diff.Differences = append(diff.Differences, FieldDiff{
				Field:          field,
				TerraformValue: tfValue,
				ActualValue:    azValue,
			})
		}
	}

	// Compare tags
	if !tagsEqual(tfRes.Attributes, azRes.Tags) {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:          "tags",
			TerraformValue: getTerraformTags(tfRes.Attributes),
			ActualValue:    azRes.Tags,
		})
	}

	return diff
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
// Azure locations and resource names are case-insensitive.
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

	// String-to-string: case-insensitive (Azure locations are case-insensitive)
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

// tagsEqual compares tags between Terraform state and Azure.
func tagsEqual(tfAttrs map[string]interface{}, azureTags map[string]string) bool {
	tfTags := getTerraformTags(tfAttrs)

	// Ignore Azure-managed tags
	ignoredPrefixes := []string{"hidden-", "ms-resource-usage"}

	filteredAzureTags := make(map[string]string)
	for k, v := range azureTags {
		ignored := false
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(strings.ToLower(k), prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			filteredAzureTags[k] = v
		}
	}

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
