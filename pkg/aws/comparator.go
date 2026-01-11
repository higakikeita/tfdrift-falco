package aws

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// CompareStateWithActual compares Terraform state with actual AWS resources
// and returns the differences (unmanaged, missing, and modified resources)
func CompareStateWithActual(tfResources []*terraform.Resource, awsResources []*DiscoveredResource) *DriftResult {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{},
		MissingResources:   []*terraform.Resource{},
		ModifiedResources:  []*ResourceDiff{},
	}

	// Build maps for efficient lookup
	tfResourceMap := make(map[string]*terraform.Resource)
	for _, tfRes := range tfResources {
		// Extract the resource ID from Terraform state
		resourceID := extractTFResourceID(tfRes)
		if resourceID != "" {
			tfResourceMap[resourceID] = tfRes
		}
	}

	awsResourceMap := make(map[string]*DiscoveredResource)
	for _, awsRes := range awsResources {
		awsResourceMap[awsRes.ID] = awsRes
	}

	log.Infof("Comparing state: %d Terraform resources vs %d AWS resources", len(tfResourceMap), len(awsResourceMap))

	// Find unmanaged resources (in AWS but not in Terraform)
	for awsID, awsRes := range awsResourceMap {
		if _, exists := tfResourceMap[awsID]; !exists {
			log.Infof("Found unmanaged resource: %s (%s)", awsID, awsRes.Type)
			result.UnmanagedResources = append(result.UnmanagedResources, awsRes)
		}
	}

	// Find missing resources (in Terraform but not in AWS) and modified resources
	for tfID, tfRes := range tfResourceMap {
		awsRes, exists := awsResourceMap[tfID]
		if !exists {
			log.Infof("Found missing resource: %s (%s)", tfID, tfRes.Type)
			result.MissingResources = append(result.MissingResources, tfRes)
		} else {
			// Resource exists in both - check for modifications
			diff := compareResourceAttributes(tfRes, awsRes)
			if diff != nil && len(diff.Differences) > 0 {
				log.Infof("Found modified resource: %s (%d differences)", tfID, len(diff.Differences))
				result.ModifiedResources = append(result.ModifiedResources, diff)
			}
		}
	}

	log.Infof("Drift detection complete: %d unmanaged, %d missing, %d modified",
		len(result.UnmanagedResources), len(result.MissingResources), len(result.ModifiedResources))

	return result
}

// extractTFResourceID extracts the AWS resource ID from Terraform resource attributes
func extractTFResourceID(tfRes *terraform.Resource) string {
	// Try common ID fields
	idFields := []string{"id", "instance_id", "db_instance_identifier", "vpc_id",
		"subnet_id", "group_id", "cluster_name", "replication_group_id", "arn"}

	for _, field := range idFields {
		if id, ok := tfRes.Attributes[field].(string); ok && id != "" {
			return id
		}
	}

	return ""
}

// compareResourceAttributes compares attributes between Terraform state and actual AWS resource
func compareResourceAttributes(tfRes *terraform.Resource, awsRes *DiscoveredResource) *ResourceDiff {
	if tfRes.Type != awsRes.Type {
		// Type mismatch - shouldn't happen if IDs match
		return nil
	}

	diff := &ResourceDiff{
		ResourceID:     awsRes.ID,
		ResourceType:   tfRes.Type,
		TerraformState: tfRes.Attributes,
		ActualState:    awsRes.Attributes,
		Differences:    []FieldDiff{},
	}

	// Compare key attributes based on resource type
	fieldsToCompare := getComparableFields(tfRes.Type)

	for _, field := range fieldsToCompare {
		tfValue := getNestedValue(tfRes.Attributes, field)
		awsValue := getNestedValue(awsRes.Attributes, field)

		if !valuesEqual(tfValue, awsValue) {
			diff.Differences = append(diff.Differences, FieldDiff{
				Field:          field,
				TerraformValue: tfValue,
				ActualValue:    awsValue,
			})
		}
	}

	// Compare tags separately
	if !tagsEqual(tfRes.Attributes, awsRes.Tags) {
		diff.Differences = append(diff.Differences, FieldDiff{
			Field:          "tags",
			TerraformValue: getTerraformTags(tfRes.Attributes),
			ActualValue:    awsRes.Tags,
		})
	}

	return diff
}

// getComparableFields returns the list of fields to compare for a given resource type
func getComparableFields(resourceType string) []string {
	switch resourceType {
	case "aws_vpc":
		return []string{"cidr_block", "enable_dns_hostnames", "enable_dns_support"}
	case "aws_subnet":
		return []string{"vpc_id", "cidr_block", "availability_zone", "map_public_ip_on_launch"}
	case "aws_security_group":
		return []string{"vpc_id", "description", "name"}
	case "aws_instance":
		return []string{"instance_type", "subnet_id", "vpc_id", "availability_zone"}
	case "aws_db_instance":
		return []string{"engine", "engine_version", "instance_class", "allocated_storage",
			"db_subnet_group_name", "multi_az", "publicly_accessible"}
	case "aws_eks_cluster":
		return []string{"version", "role_arn"}
	case "aws_elasticache_replication_group":
		return []string{"node_type", "automatic_failover_enabled", "multi_az_enabled"}
	case "aws_lb":
		return []string{"type", "scheme", "vpc_id"}
	default:
		return []string{}
	}
}

// getNestedValue retrieves a value from a nested map using dot notation (e.g., "vpc_config.subnet_ids")
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

		// Navigate deeper
		if nextMap, ok := value.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return nil
		}
	}

	return nil
}

// valuesEqual compares two values for equality, handling different types
func valuesEqual(a, b interface{}) bool {
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
			return aBool == true
		}
		if bStr == "false" {
			return aBool == false
		}
	}

	// Handle numeric comparisons
	if reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind() {
		return reflect.DeepEqual(a, b)
	}

	// Fallback to string comparison
	return aStr == bStr
}

// tagsEqual compares tags between Terraform state and AWS
func tagsEqual(tfAttrs map[string]interface{}, awsTags map[string]string) bool {
	tfTags := getTerraformTags(tfAttrs)

	// Ignore AWS-managed tags
	ignoredPrefixes := []string{"aws:", "kubernetes.io/"}

	filteredAwsTags := make(map[string]string)
	for k, v := range awsTags {
		ignored := false
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(k, prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			filteredAwsTags[k] = v
		}
	}

	return reflect.DeepEqual(tfTags, filteredAwsTags)
}

// getTerraformTags extracts tags from Terraform resource attributes
func getTerraformTags(attrs map[string]interface{}) map[string]string {
	tags := make(map[string]string)

	// Try "tags" field (most common)
	if tagsVal, ok := attrs["tags"].(map[string]interface{}); ok {
		for k, v := range tagsVal {
			if strVal, ok := v.(string); ok {
				tags[k] = strVal
			}
		}
		return tags
	}

	// Try "tags_all" field (Terraform AWS provider v4+)
	if tagsAllVal, ok := attrs["tags_all"].(map[string]interface{}); ok {
		for k, v := range tagsAllVal {
			if strVal, ok := v.(string); ok {
				tags[k] = strVal
			}
		}
		return tags
	}

	return tags
}
