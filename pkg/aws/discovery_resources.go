package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// discoverRDSInstances discovers all RDS instances in the region
func (d *DiscoveryClient) discoverRDSInstances(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS Instances: %w", err)
	}

	var resources []*DiscoveredResource
	for _, db := range result.DBInstances {
		tags := extractRDSTags(db.TagList)

		var securityGroupIDs []string
		for _, sg := range db.VpcSecurityGroups {
			securityGroupIDs = append(securityGroupIDs, aws.ToString(sg.VpcSecurityGroupId))
		}

		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(db.DBInstanceIdentifier),
			Type:   "aws_db_instance",
			ARN:    aws.ToString(db.DBInstanceArn),
			Name:   aws.ToString(db.DBInstanceIdentifier),
			Region: d.region,
			Attributes: map[string]interface{}{
				"engine":                  aws.ToString(db.Engine),
				"engine_version":          aws.ToString(db.EngineVersion),
				"instance_class":          aws.ToString(db.DBInstanceClass),
				"allocated_storage":       db.AllocatedStorage,
				"db_subnet_group_name":    aws.ToString(db.DBSubnetGroup.DBSubnetGroupName),
				"vpc_security_group_ids":  securityGroupIDs,
				"availability_zone":       aws.ToString(db.AvailabilityZone),
				"multi_az":                db.MultiAZ,
				"publicly_accessible":     db.PubliclyAccessible,
				"status":                  aws.ToString(db.DBInstanceStatus),
			},
			Tags: tags,
		})
	}

	return resources, nil
}

// discoverEKSClusters discovers all EKS clusters in the region
func (d *DiscoveryClient) discoverEKSClusters(ctx context.Context) ([]*DiscoveredResource, error) {
	// First, list all cluster names
	listResult, err := d.eksClient.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS Clusters: %w", err)
	}

	var resources []*DiscoveredResource
	for _, clusterName := range listResult.Clusters {
		// Then describe each cluster to get details
		descResult, err := d.eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe EKS Cluster %s: %w", clusterName, err)
		}

		cluster := descResult.Cluster
		tags := cluster.Tags

		var subnetIDs []string
		if cluster.ResourcesVpcConfig != nil {
			for _, subnet := range cluster.ResourcesVpcConfig.SubnetIds {
				subnetIDs = append(subnetIDs, subnet)
			}
		}

		var securityGroupIDs []string
		if cluster.ResourcesVpcConfig != nil {
			for _, sg := range cluster.ResourcesVpcConfig.SecurityGroupIds {
				securityGroupIDs = append(securityGroupIDs, sg)
			}
		}

		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(cluster.Name),
			Type:   "aws_eks_cluster",
			ARN:    aws.ToString(cluster.Arn),
			Name:   aws.ToString(cluster.Name),
			Region: d.region,
			Attributes: map[string]interface{}{
				"version":            aws.ToString(cluster.Version),
				"role_arn":           aws.ToString(cluster.RoleArn),
				"status":             string(cluster.Status),
				"subnet_ids":         subnetIDs,
				"security_group_ids": securityGroupIDs,
				"endpoint":           aws.ToString(cluster.Endpoint),
			},
			Tags: tags,
		})
	}

	return resources, nil
}

// discoverElastiCacheClusters discovers all ElastiCache replication groups in the region
func (d *DiscoveryClient) discoverElastiCacheClusters(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.elasticache.DescribeReplicationGroups(ctx, &elasticache.DescribeReplicationGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe ElastiCache Replication Groups: %w", err)
	}

	var resources []*DiscoveredResource
	for _, rg := range result.ReplicationGroups {
		var nodeTypes []string
		for _, nodeGroup := range rg.NodeGroups {
			for _, member := range nodeGroup.NodeGroupMembers {
				if member.CacheNodeId != nil {
					nodeTypes = append(nodeTypes, aws.ToString(member.CacheNodeId))
				}
			}
		}

		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(rg.ReplicationGroupId),
			Type:   "aws_elasticache_replication_group",
			ARN:    aws.ToString(rg.ARN),
			Name:   aws.ToString(rg.ReplicationGroupId),
			Region: d.region,
			Attributes: map[string]interface{}{
				"description":                aws.ToString(rg.Description),
				"node_type":                  aws.ToString(rg.CacheNodeType),
				"num_cache_clusters":         len(rg.MemberClusters),
				"automatic_failover_enabled": string(rg.AutomaticFailover),
				"multi_az_enabled":           string(rg.MultiAZ),
				"status":                     aws.ToString(rg.Status),
			},
		})
	}

	return resources, nil
}

// discoverLoadBalancers discovers all Application/Network Load Balancers in the region
func (d *DiscoveryClient) discoverLoadBalancers(ctx context.Context) ([]*DiscoveredResource, error) {
	result, err := d.elbClient.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Load Balancers: %w", err)
	}

	var resources []*DiscoveredResource
	for _, lb := range result.LoadBalancers {
		// Get tags for this load balancer
		tagsResult, err := d.elbClient.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{aws.ToString(lb.LoadBalancerArn)},
		})

		var tags map[string]string
		if err == nil && len(tagsResult.TagDescriptions) > 0 {
			tags = extractELBTags(tagsResult.TagDescriptions[0].Tags)
		}

		var subnetIDs []string
		for _, az := range lb.AvailabilityZones {
			subnetIDs = append(subnetIDs, aws.ToString(az.SubnetId))
		}

		var securityGroupIDs []string
		for _, sg := range lb.SecurityGroups {
			securityGroupIDs = append(securityGroupIDs, sg)
		}

		resources = append(resources, &DiscoveredResource{
			ID:     aws.ToString(lb.LoadBalancerArn),
			Type:   "aws_lb",
			ARN:    aws.ToString(lb.LoadBalancerArn),
			Name:   aws.ToString(lb.LoadBalancerName),
			Region: d.region,
			Attributes: map[string]interface{}{
				"type":               string(lb.Type),
				"scheme":             string(lb.Scheme),
				"vpc_id":             aws.ToString(lb.VpcId),
				"subnets":            subnetIDs,
				"security_groups":    securityGroupIDs,
				"dns_name":           aws.ToString(lb.DNSName),
				"state":              string(lb.State.Code),
			},
			Tags: tags,
		})
	}

	return resources, nil
}

// Helper functions for different AWS service tag formats

func extractRDSTags(tags []rdsTypes.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

func extractELBTags(tags []elbTypes.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}
