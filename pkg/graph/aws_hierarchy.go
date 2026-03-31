package graph

import (
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// AWSHierarchyBuilder builds AWS standard hierarchical structure
type AWSHierarchyBuilder struct {
	resources map[string]*terraform.Resource
	hierarchy *AWSHierarchy
}

// AWSHierarchy represents AWS infrastructure hierarchy
type AWSHierarchy struct {
	Regions map[string]*Region
}

// Region represents an AWS region
type Region struct {
	ID   string
	Name string
	VPCs map[string]*VPC
}

// VPC represents a Virtual Private Cloud
type VPC struct {
	ID                string
	Name              string
	CIDR              string
	AvailabilityZones map[string]*AvailabilityZone
	Resources         []string // Resources not in subnets
}

// AvailabilityZone represents an AZ within a VPC
type AvailabilityZone struct {
	ID      string
	Name    string
	Subnets map[string]*Subnet
}

// Subnet represents a subnet within an AZ
type Subnet struct {
	ID        string
	Name      string
	CIDR      string
	Type      string // "public" or "private"
	Resources []string
}

// NewAWSHierarchyBuilder creates a new AWS hierarchy builder
func NewAWSHierarchyBuilder() *AWSHierarchyBuilder {
	return &AWSHierarchyBuilder{
		resources: make(map[string]*terraform.Resource),
		hierarchy: &AWSHierarchy{
			Regions: make(map[string]*Region),
		},
	}
}

// Build builds hierarchical structure from Terraform resources
func (hb *AWSHierarchyBuilder) Build(resources []*terraform.Resource) *AWSHierarchy {
	log.Info("Building AWS hierarchy structure...")

	// Index resources by ID
	for _, res := range resources {
		id := extractResourceIDFromAttributes(res.Attributes)
		if id != "" {
			hb.resources[id] = res
		}
	}

	// First pass: Build VPC structure
	for _, res := range resources {
		switch res.Type {
		case "aws_vpc":
			hb.addVPC(res)
		}
	}

	// Second pass: Build Subnet structure
	for _, res := range resources {
		switch res.Type {
		case "aws_subnet":
			hb.addSubnet(res)
		}
	}

	// Third pass: Assign resources to subnets/VPCs
	for _, res := range resources {
		switch res.Type {
		case "aws_instance", "aws_db_instance", "aws_nat_gateway":
			hb.assignResourceToSubnet(res)
		case "aws_internet_gateway", "aws_route_table", "aws_security_group":
			hb.assignResourceToVPC(res)
		}
	}

	log.Infof("Built hierarchy: %d regions, total VPCs across regions", len(hb.hierarchy.Regions))
	return hb.hierarchy
}

// addVPC adds a VPC to the hierarchy
func (hb *AWSHierarchyBuilder) addVPC(res *terraform.Resource) {
	vpcID := extractResourceIDFromAttributes(res.Attributes)
	if vpcID == "" {
		return
	}

	// Extract region from provider or attributes
	region := hb.extractRegion(res)
	if region == "" {
		region = "us-east-1" // Default region
	}

	// Ensure region exists
	if _, exists := hb.hierarchy.Regions[region]; !exists {
		hb.hierarchy.Regions[region] = &Region{
			ID:   region,
			Name: region,
			VPCs: make(map[string]*VPC),
		}
	}

	// Extract VPC details
	vpcName := hb.extractResourceName(res)
	cidr := ""
	if cidrBlock, ok := res.Attributes["cidr_block"].(string); ok {
		cidr = cidrBlock
	}

	hb.hierarchy.Regions[region].VPCs[vpcID] = &VPC{
		ID:                vpcID,
		Name:              vpcName,
		CIDR:              cidr,
		AvailabilityZones: make(map[string]*AvailabilityZone),
		Resources:         []string{},
	}

	log.Debugf("Added VPC: %s (%s) in region %s", vpcName, cidr, region)
}

// addSubnet adds a subnet to the hierarchy
func (hb *AWSHierarchyBuilder) addSubnet(res *terraform.Resource) {
	subnetID := extractResourceIDFromAttributes(res.Attributes)
	if subnetID == "" {
		return
	}

	// Get VPC ID
	vpcID, ok := res.Attributes["vpc_id"].(string)
	if !ok || vpcID == "" {
		return
	}

	// Find VPC in hierarchy
	var vpc *VPC
	var region string
	for r, reg := range hb.hierarchy.Regions {
		if v, exists := reg.VPCs[vpcID]; exists {
			vpc = v
			region = r
			break
		}
	}

	if vpc == nil {
		log.Warnf("VPC not found for subnet %s", subnetID)
		return
	}

	// Extract AZ
	az := ""
	if availabilityZone, ok := res.Attributes["availability_zone"].(string); ok {
		az = availabilityZone
	}
	if az == "" {
		az = region + "a" // Default AZ
	}

	// Ensure AZ exists in VPC
	if _, exists := vpc.AvailabilityZones[az]; !exists {
		vpc.AvailabilityZones[az] = &AvailabilityZone{
			ID:      az,
			Name:    az,
			Subnets: make(map[string]*Subnet),
		}
	}

	// Extract subnet details
	subnetName := hb.extractResourceName(res)
	cidr := ""
	if cidrBlock, ok := res.Attributes["cidr_block"].(string); ok {
		cidr = cidrBlock
	}

	// Determine subnet type (public/private)
	subnetType := "private" // Default to private
	if mapPublicIP, ok := res.Attributes["map_public_ip_on_launch"].(bool); ok && mapPublicIP {
		subnetType = "public"
	}

	vpc.AvailabilityZones[az].Subnets[subnetID] = &Subnet{
		ID:        subnetID,
		Name:      subnetName,
		CIDR:      cidr,
		Type:      subnetType,
		Resources: []string{},
	}

	log.Debugf("Added subnet: %s (%s) [%s] in AZ %s", subnetName, cidr, subnetType, az)
}

// assignResourceToSubnet assigns a resource to its subnet
func (hb *AWSHierarchyBuilder) assignResourceToSubnet(res *terraform.Resource) {
	resourceID := extractResourceIDFromAttributes(res.Attributes)
	if resourceID == "" {
		return
	}

	// Get subnet ID
	var subnetID string
	switch res.Type {
	case "aws_instance":
		if subnet, ok := res.Attributes["subnet_id"].(string); ok {
			subnetID = subnet
		}
	case "aws_db_instance":
		// RDS instances use subnet groups
		// For simplicity, try to extract from subnet_group_name
		if subnetGroupName, ok := res.Attributes["db_subnet_group_name"].(string); ok {
			log.Debugf("RDS instance %s in subnet group %s", resourceID, subnetGroupName)
		}
	case "aws_nat_gateway":
		if subnet, ok := res.Attributes["subnet_id"].(string); ok {
			subnetID = subnet
		}
	}

	if subnetID == "" {
		return
	}

	// Find subnet in hierarchy
	for _, region := range hb.hierarchy.Regions {
		for _, vpc := range region.VPCs {
			for _, az := range vpc.AvailabilityZones {
				if subnet, exists := az.Subnets[subnetID]; exists {
					subnet.Resources = append(subnet.Resources, resourceID)
					log.Debugf("Assigned resource %s to subnet %s", resourceID, subnetID)
					return
				}
			}
		}
	}

	log.Warnf("Subnet %s not found for resource %s", subnetID, resourceID)
}

// assignResourceToVPC assigns a resource to its VPC
func (hb *AWSHierarchyBuilder) assignResourceToVPC(res *terraform.Resource) {
	resourceID := extractResourceIDFromAttributes(res.Attributes)
	if resourceID == "" {
		return
	}

	// Get VPC ID
	vpcID, ok := res.Attributes["vpc_id"].(string)
	if !ok || vpcID == "" {
		return
	}

	// Find VPC in hierarchy
	for _, region := range hb.hierarchy.Regions {
		if vpc, exists := region.VPCs[vpcID]; exists {
			vpc.Resources = append(vpc.Resources, resourceID)
			log.Debugf("Assigned resource %s to VPC %s", resourceID, vpcID)
			return
		}
	}

	log.Warnf("VPC %s not found for resource %s", vpcID, resourceID)
}

// extractRegion extracts region from resource
func (hb *AWSHierarchyBuilder) extractRegion(res *terraform.Resource) string {
	// Try to extract from provider
	if provider, ok := res.Attributes["provider"].(string); ok {
		// Provider format: provider["registry.terraform.io/hashicorp/aws"].us-east-1
		if strings.Contains(provider, ".") {
			parts := strings.Split(provider, ".")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	// Try common region attributes
	if region, ok := res.Attributes["region"].(string); ok {
		return region
	}

	// Parse from ARN if available
	if arn, ok := res.Attributes["arn"].(string); ok {
		// ARN format: arn:aws:service:region:account-id:resource
		parts := strings.Split(arn, ":")
		if len(parts) >= 4 {
			return parts[3]
		}
	}

	return ""
}

// extractResourceName extracts a human-readable name from resource
func (hb *AWSHierarchyBuilder) extractResourceName(res *terraform.Resource) string {
	// Try name attribute
	if name, ok := res.Attributes["name"].(string); ok && name != "" {
		return name
	}

	// Try tags.Name
	if tags, ok := res.Attributes["tags"].(map[string]interface{}); ok {
		if name, ok := tags["Name"].(string); ok && name != "" {
			return name
		}
	}

	// Fallback to Terraform resource name
	if res.Name != "" {
		return res.Name
	}

	return res.Type
}
