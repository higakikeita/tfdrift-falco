package aws

import (
	"reflect"
	"sort"

	"github.com/higakikeita/driftwire/pkg/comparator"
	"github.com/higakikeita/driftwire/pkg/terraform"
	"github.com/higakikeita/driftwire/pkg/types"
)

// discoverableResourceTypes is the set of AWS resource types DiscoverAll actually
// queries. A Terraform resource is only reported as `missing` when its type is in
// this set — otherwise "not found in the cloud scan" just means "we never scanned
// that type", which was being mis-reported as drift and inflating `missing` (#338).
var discoverableResourceTypes = map[string]bool{
	"aws_vpc":                           true,
	"aws_subnet":                        true,
	"aws_security_group":                true,
	"aws_instance":                      true,
	"aws_db_instance":                   true,
	"aws_eks_cluster":                   true,
	"aws_elasticache_replication_group": true,
	"aws_lb":                            true,
}

// CompareStateWithActual compares Terraform state with actual AWS resources
// and returns the differences (unmanaged, missing, and modified resources)
func CompareStateWithActual(tfResources []*terraform.Resource, awsResources []*types.DiscoveredResource) *types.DriftResult {
	config := &comparator.ComparisonConfig{
		ExtractTFID: extractTFResourceID,
		ExtractCloudID: func(cloudResource interface{}) string {
			res, _ := cloudResource.(*types.DiscoveredResource)
			if res == nil {
				return ""
			}
			return res.ID
		},
		CompareAttributes: func(tfResource, cloudResource interface{}) []types.FieldDiff {
			tfRes, _ := tfResource.(*terraform.Resource)
			cloudRes, _ := cloudResource.(*types.DiscoveredResource)
			if tfRes == nil || cloudRes == nil {
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
			awsRes, _ := cloudResource.(*types.DiscoveredResource)
			if awsRes == nil {
				return nil
			}
			return &types.ResourceDiff{
				ResourceID:   awsRes.ID,
				ResourceType: awsRes.Type,
				Provider:     "aws",
				ActualState:  awsRes.Attributes,
				Differences:  []types.FieldDiff{},
			}
		},
		BuildMissing: func(tfResource interface{}) *types.TerraformResource {
			tfRes, _ := tfResource.(*terraform.Resource)
			if tfRes == nil {
				return nil
			}
			return &types.TerraformResource{
				Type:       tfRes.Type,
				Name:       tfRes.Name,
				ID:         extractTFResourceID(tfRes),
				Provider:   "aws",
				Attributes: tfRes.Attributes,
			}
		},
	}

	result := comparator.CompareResources(config, convertTFResources(tfResources), convertCloudResources(awsResources))
	result.Provider = "aws"

	// Scope `missing` to types discovery actually scans. A resource of an
	// unscanned type has no cloud match not because it was deleted but because we
	// never looked — reporting it as missing is a false positive (#338).
	scoped := result.MissingResources[:0]
	for _, m := range result.MissingResources {
		if m != nil && discoverableResourceTypes[m.Type] {
			scoped = append(scoped, m)
		}
	}
	result.MissingResources = scoped

	return result
}

// convertTFResources converts []*terraform.Resource to []interface{}
func convertTFResources(resources []*terraform.Resource) []interface{} {
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

// extractTFResourceID extracts the AWS resource ID from Terraform resource attributes
func extractTFResourceID(resource interface{}) string {
	tfRes, _ := resource.(*terraform.Resource)
	if tfRes == nil {
		return ""
	}
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
func compareResourceAttributes(tfRes *terraform.Resource, awsRes *types.DiscoveredResource) []*types.FieldDiff {
	if tfRes.Type != awsRes.Type {
		// Type mismatch - shouldn't happen if IDs match
		return nil
	}

	// Compare key attributes based on resource type
	fieldsToCompare := getComparableFields(tfRes.Type)
	differences := []*types.FieldDiff{}

	for _, field := range fieldsToCompare {
		tfValue := comparator.GetNestedValue(tfRes.Attributes, field)
		awsValue := comparator.GetNestedValue(awsRes.Attributes, field)

		if !comparator.ValuesEqual(tfValue, awsValue) {
			differences = append(differences, &types.FieldDiff{
				Field:          field,
				TerraformValue: tfValue,
				ActualValue:    awsValue,
			})
		}
	}

	// Security-group rules: the AWS side is a pre-canonicalized rule set (from
	// discovery); canonicalize the Terraform blocks the same way and compare as
	// unordered sets, so an added/removed ingress/egress rule is caught (#338).
	if tfRes.Type == "aws_security_group" {
		for _, field := range []string{"ingress", "egress"} {
			tfRules := canonicalizeTFRules(tfRes.Attributes[field])
			awsRules := toStringSliceVal(awsRes.Attributes[field])
			sort.Strings(awsRules)
			// Treat nil and empty as equal (a missing key is not drift); only a
			// genuine set difference counts.
			if len(tfRules) == 0 && len(awsRules) == 0 {
				continue
			}
			if !reflect.DeepEqual(tfRules, awsRules) {
				differences = append(differences, &types.FieldDiff{
					Field:          field,
					TerraformValue: tfRules,
					ActualValue:    awsRules,
				})
			}
		}
	}

	// Compare tags separately
	if !tagsEqual(tfRes.Attributes, awsRes.Tags) {
		differences = append(differences, &types.FieldDiff{
			Field:          "tags",
			TerraformValue: getTerraformTags(tfRes.Attributes),
			ActualValue:    awsRes.Tags,
		})
	}

	return differences
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

// tagsEqual compares tags between Terraform state and AWS (provider-specific logic)
func tagsEqual(tfAttrs map[string]interface{}, awsTags map[string]string) bool {
	tfTags := getTerraformTags(tfAttrs)

	// Ignore AWS-managed tags
	ignoredPrefixes := []string{"aws:", "kubernetes.io/"}

	filteredAwsTags := comparator.FilterManagedLabels(awsTags, ignoredPrefixes)

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
