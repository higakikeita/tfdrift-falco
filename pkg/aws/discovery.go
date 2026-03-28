package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	log "github.com/sirupsen/logrus"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// DiscoveryClient handles AWS resource discovery
type DiscoveryClient struct {
	region         string
	ec2Client      *ec2.Client
	rdsClient      *rds.Client
	eksClient      *eks.Client
	elasticache    *elasticache.Client
	elbClient      *elasticloadbalancingv2.Client
}

// DiscoveredResource represents a resource found in AWS
type DiscoveredResource struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	ARN        string                 `json:"arn,omitempty"`
	Name       string                 `json:"name"`
	Region     string                 `json:"region"`
	Attributes map[string]interface{} `json:"attributes"`
	Tags       map[string]string      `json:"tags,omitempty"`
}

// DriftResult represents the difference between Terraform and actual AWS state
type DriftResult struct {
	// Resources in AWS but not in Terraform (manually created)
	UnmanagedResources []*DiscoveredResource `json:"unmanaged_resources"`

	// Resources in Terraform but not in AWS (manually deleted)
	MissingResources []*terraform.Resource `json:"missing_resources"`

	// Resources with configuration differences
	ModifiedResources []*ResourceDiff `json:"modified_resources"`
}

// ResourceDiff represents differences in a single resource
type ResourceDiff struct {
	ResourceID         string                 `json:"resource_id"`
	ResourceType       string                 `json:"resource_type"`
	TerraformState     map[string]interface{} `json:"terraform_state"`
	ActualState        map[string]interface{} `json:"actual_state"`
	Differences        []FieldDiff            `json:"differences"`
}

// FieldDiff represents a difference in a specific field
type FieldDiff struct {
	Field          string      `json:"field"`
	TerraformValue interface{} `json:"terraform_value"`
	ActualValue    interface{} `json:"actual_value"`
}

// NewDiscoveryClient creates a new AWS discovery client
func NewDiscoveryClient(ctx context.Context, region string) (*DiscoveryClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &DiscoveryClient{
		region:      region,
		ec2Client:   ec2.NewFromConfig(cfg),
		rdsClient:   rds.NewFromConfig(cfg),
		eksClient:   eks.NewFromConfig(cfg),
		elasticache: elasticache.NewFromConfig(cfg),
		elbClient:   elasticloadbalancingv2.NewFromConfig(cfg),
	}, nil
}

// DiscoverAll discovers all supported AWS resources in the region
func (d *DiscoveryClient) DiscoverAll(ctx context.Context) ([]*DiscoveredResource, error) {
	log.Infof("Starting AWS resource discovery in region %s", d.region)

	var allResources []*DiscoveredResource

	// Discover VPCs
	vpcs, err := d.discoverVPCs(ctx)
	if err != nil {
		log.Warnf("Failed to discover VPCs: %v", err)
	} else {
		allResources = append(allResources, vpcs...)
		log.Infof("Discovered %d VPCs", len(vpcs))
	}

	// Discover Subnets
	subnets, err := d.discoverSubnets(ctx)
	if err != nil {
		log.Warnf("Failed to discover Subnets: %v", err)
	} else {
		allResources = append(allResources, subnets...)
		log.Infof("Discovered %d Subnets", len(subnets))
	}

	// Discover Security Groups
	sgs, err := d.discoverSecurityGroups(ctx)
	if err != nil {
		log.Warnf("Failed to discover Security Groups: %v", err)
	} else {
		allResources = append(allResources, sgs...)
		log.Infof("Discovered %d Security Groups", len(sgs))
	}

	// Discover EC2 Instances
	instances, err := d.discoverEC2Instances(ctx)
	if err != nil {
		log.Warnf("Failed to discover EC2 Instances: %v", err)
	} else {
		allResources = append(allResources, instances...)
		log.Infof("Discovered %d EC2 Instances", len(instances))
	}

	// Discover RDS Instances
	rdsInstances, err := d.discoverRDSInstances(ctx)
	if err != nil {
		log.Warnf("Failed to discover RDS Instances: %v", err)
	} else {
		allResources = append(allResources, rdsInstances...)
		log.Infof("Discovered %d RDS Instances", len(rdsInstances))
	}

	// Discover EKS Clusters
	eksClusters, err := d.discoverEKSClusters(ctx)
	if err != nil {
		log.Warnf("Failed to discover EKS Clusters: %v", err)
	} else {
		allResources = append(allResources, eksClusters...)
		log.Infof("Discovered %d EKS Clusters", len(eksClusters))
	}

	// Discover ElastiCache Clusters
	caches, err := d.discoverElastiCacheClusters(ctx)
	if err != nil {
		log.Warnf("Failed to discover ElastiCache Clusters: %v", err)
	} else {
		allResources = append(allResources, caches...)
		log.Infof("Discovered %d ElastiCache Clusters", len(caches))
	}

	// Discover Load Balancers
	lbs, err := d.discoverLoadBalancers(ctx)
	if err != nil {
		log.Warnf("Failed to discover Load Balancers: %v", err)
	} else {
		allResources = append(allResources, lbs...)
		log.Infof("Discovered %d Load Balancers", len(lbs))
	}

	log.Infof("AWS discovery completed: %d total resources discovered", len(allResources))
	return allResources, nil
}

// discoverVPCs discovers all VPCs in the region
func (d *DiscoveryClient) discoverVPCs(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	var resources []*DiscoveredResource
	for _, vpc := range result.Vpcs {
		tags := extractTags(vpc.Tags)
		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(vpc.VpcId),
			Type:   "aws_vpc",
			Name:   getTagValue(vpc.Tags, "Name"),
			Region: d.region,
			Attributes: map[string]interface{}{
				"cidr_block": aws.ToString(vpc.CidrBlock),
				"state":      string(vpc.State),
			},
			Tags: tags,
		})
	}

	return resources, nil
}

// discoverSubnets discovers all subnets in the region
func (d *DiscoveryClient) discoverSubnets(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Subnets: %w", err)
	}

	var resources []*DiscoveredResource
	for _, subnet := range result.Subnets {
		tags := extractTags(subnet.Tags)
		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(subnet.SubnetId),
			Type:   "aws_subnet",
			Name:   getTagValue(subnet.Tags, "Name"),
			Region: d.region,
			Attributes: map[string]interface{}{
				"vpc_id":                       aws.ToString(subnet.VpcId),
				"cidr_block":                   aws.ToString(subnet.CidrBlock),
				"availability_zone":            aws.ToString(subnet.AvailabilityZone),
				"map_public_ip_on_launch":      subnet.MapPublicIpOnLaunch,
				"assign_ipv6_address_on_creation": subnet.AssignIpv6AddressOnCreation,
			},
			Tags: tags,
		})
	}

	return resources, nil
}

// discoverSecurityGroups discovers all security groups in the region
func (d *DiscoveryClient) discoverSecurityGroups(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Security Groups: %w", err)
	}

	var resources []*DiscoveredResource
	for _, sg := range result.SecurityGroups {
		tags := extractTags(sg.Tags)
		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(sg.GroupId),
			Type:   "aws_security_group",
			Name:   aws.ToString(sg.GroupName),
			Region: d.region,
			Attributes: map[string]interface{}{
				"vpc_id":      aws.ToString(sg.VpcId),
				"description": aws.ToString(sg.Description),
				"name":        aws.ToString(sg.GroupName),
			},
			Tags: tags,
		})
	}

	return resources, nil
}

// discoverEC2Instances discovers all EC2 instances in the region
func (d *DiscoveryClient) discoverEC2Instances(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 Instances: %w", err)
	}

	var resources []*DiscoveredResource
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			tags := extractTags(instance.Tags)
			resources = append(resources, &DiscoveredResource{
				ID:     aws.ToString(instance.InstanceId),
				Type:   "aws_instance",
				Name:   getTagValue(instance.Tags, "Name"),
				Region: d.region,
				Attributes: map[string]interface{}{
					"instance_type":      string(instance.InstanceType),
					"subnet_id":          aws.ToString(instance.SubnetId),
					"vpc_id":             aws.ToString(instance.VpcId),
					"availability_zone":  aws.ToString(instance.Placement.AvailabilityZone),
					"private_ip":         aws.ToString(instance.PrivateIpAddress),
					"public_ip":          aws.ToString(instance.PublicIpAddress),
					"state":              string(instance.State.Name),
				},
				Tags: tags,
			})
		}
	}

	return resources, nil
}

// Helper functions

func extractTags(tags []ec2Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

func getTagValue(tags []ec2Types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}
