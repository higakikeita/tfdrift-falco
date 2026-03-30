package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eksTypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// TestDiscoverVPCs_Success tests successful VPC discovery
func TestDiscoverVPCs_Success(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeVpcsFunc: func(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []ec2Types.Vpc{
					{
						VpcId:     aws.String("vpc-12345"),
						CidrBlock: aws.String("10.0.0.0/16"),
						State:     ec2Types.VpcStateAvailable,
						Tags: []ec2Types.Tag{
							{Key: aws.String("Name"), Value: aws.String("prod-vpc")},
						},
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	resources, err := client.discoverVPCs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 VPC, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "vpc-12345" {
		t.Errorf("expected ID vpc-12345, got %s", resource.ID)
	}
	if resource.Type != "aws_vpc" {
		t.Errorf("expected Type aws_vpc, got %s", resource.Type)
	}
	if resource.Name != "prod-vpc" {
		t.Errorf("expected Name prod-vpc, got %s", resource.Name)
	}
}

// TestDiscoverVPCs_Error tests VPC discovery error handling
func TestDiscoverVPCs_Error(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeVpcsFunc: func(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
			return nil, errors.New("access denied")
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	_, err := client.discoverVPCs(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestDiscoverSubnets_Success tests successful subnet discovery
func TestDiscoverSubnets_Success(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeSubnetsFunc: func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
			return &ec2.DescribeSubnetsOutput{
				Subnets: []ec2Types.Subnet{
					{
						SubnetId:            aws.String("subnet-12345"),
						VpcId:               aws.String("vpc-12345"),
						CidrBlock:           aws.String("10.0.1.0/24"),
						AvailabilityZone:    aws.String("us-east-1a"),
						MapPublicIpOnLaunch: aws.Bool(true),
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	resources, err := client.discoverSubnets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 subnet, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "subnet-12345" {
		t.Errorf("expected ID subnet-12345, got %s", resource.ID)
	}
	if resource.Type != "aws_subnet" {
		t.Errorf("expected Type aws_subnet, got %s", resource.Type)
	}
}

// TestDiscoverSecurityGroups_Success tests successful security group discovery
func TestDiscoverSecurityGroups_Success(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeSecurityGroupsFunc: func(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2Types.SecurityGroup{
					{
						GroupId:     aws.String("sg-12345"),
						GroupName:   aws.String("prod-sg"),
						VpcId:       aws.String("vpc-12345"),
						Description: aws.String("Production security group"),
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	resources, err := client.discoverSecurityGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 security group, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "sg-12345" {
		t.Errorf("expected ID sg-12345, got %s", resource.ID)
	}
	if resource.Type != "aws_security_group" {
		t.Errorf("expected Type aws_security_group, got %s", resource.Type)
	}
}

// TestDiscoverEC2Instances_Success tests successful EC2 instance discovery
func TestDiscoverEC2Instances_Success(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []ec2Types.Reservation{
					{
						Instances: []ec2Types.Instance{
							{
								InstanceId:       aws.String("i-12345"),
								InstanceType:     ec2Types.InstanceTypeT3Micro,
								SubnetId:         aws.String("subnet-12345"),
								VpcId:            aws.String("vpc-12345"),
								PrivateIpAddress: aws.String("10.0.1.10"),
								PublicIpAddress:  aws.String("52.1.1.1"),
								State: &ec2Types.InstanceState{
									Name: ec2Types.InstanceStateNameRunning,
								},
								Placement: &ec2Types.Placement{
									AvailabilityZone: aws.String("us-east-1a"),
								},
							},
						},
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	resources, err := client.discoverEC2Instances(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "i-12345" {
		t.Errorf("expected ID i-12345, got %s", resource.ID)
	}
	if resource.Type != "aws_instance" {
		t.Errorf("expected Type aws_instance, got %s", resource.Type)
	}
}

// TestDiscoverRDSInstances_Success tests successful RDS instance discovery
func TestDiscoverRDSInstances_Success(t *testing.T) {
	mockRDS := &MockRDS{
		DescribeDBInstancesFunc: func(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
			return &rds.DescribeDBInstancesOutput{
				DBInstances: []rdsTypes.DBInstance{
					{
						DBInstanceIdentifier: aws.String("prod-db"),
						DBInstanceArn:        aws.String("arn:aws:rds:us-east-1:123456789:db:prod-db"),
						Engine:               aws.String("postgres"),
						EngineVersion:        aws.String("13.7"),
						DBInstanceClass:      aws.String("db.t3.micro"),
						AllocatedStorage:     aws.Int32(100),
						DBInstanceStatus:     aws.String("available"),
						AvailabilityZone:     aws.String("us-east-1a"),
						MultiAZ:              aws.Bool(false),
						PubliclyAccessible:   aws.Bool(false),
						DBSubnetGroup: &rdsTypes.DBSubnetGroup{
							DBSubnetGroupName: aws.String("default"),
						},
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, mockRDS, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	resources, err := client.discoverRDSInstances(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 RDS instance, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "prod-db" {
		t.Errorf("expected ID prod-db, got %s", resource.ID)
	}
	if resource.Type != "aws_db_instance" {
		t.Errorf("expected Type aws_db_instance, got %s", resource.Type)
	}
}

// TestDiscoverEKSClusters_Success tests successful EKS cluster discovery
func TestDiscoverEKSClusters_Success(t *testing.T) {
	mockEKS := &MockEKS{
		ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"prod-cluster"},
			}, nil
		},
		DescribeClusterFunc: func(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
			return &eks.DescribeClusterOutput{
				Cluster: &eksTypes.Cluster{
					Name:     aws.String("prod-cluster"),
					Arn:      aws.String("arn:aws:eks:us-east-1:123456789:cluster/prod-cluster"),
					Status:   eksTypes.ClusterStatusActive,
					Version:  aws.String("1.24"),
					RoleArn:  aws.String("arn:aws:iam::123456789:role/eks-role"),
					Endpoint: aws.String("https://example.eks.amazonaws.com"),
					ResourcesVpcConfig: &eksTypes.VpcConfigResponse{
						SubnetIds:        []string{"subnet-12345", "subnet-67890"},
						SecurityGroupIds: []string{"sg-12345"},
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, mockEKS, &MockElastiCache{}, &MockELB{})

	resources, err := client.discoverEKSClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 EKS cluster, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "prod-cluster" {
		t.Errorf("expected ID prod-cluster, got %s", resource.ID)
	}
	if resource.Type != "aws_eks_cluster" {
		t.Errorf("expected Type aws_eks_cluster, got %s", resource.Type)
	}
}

// TestDiscoverElastiCacheClusters_Success tests successful ElastiCache cluster discovery
func TestDiscoverElastiCacheClusters_Success(t *testing.T) {
	mockElastiCache := &MockElastiCache{
		DescribeReplicationGroupsFunc: func(ctx context.Context, params *elasticache.DescribeReplicationGroupsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
			return &elasticache.DescribeReplicationGroupsOutput{
				ReplicationGroups: []elasticacheTypes.ReplicationGroup{
					{
						ReplicationGroupId: aws.String("prod-redis"),
						ARN:                aws.String("arn:aws:elasticache:us-east-1:123456789:replicationgroup:prod-redis"),
						Description:        aws.String("Production Redis cluster"),
						CacheNodeType:      aws.String("cache.t3.micro"),
						AutomaticFailover:  elasticacheTypes.AutomaticFailoverStatusEnabled,
						MultiAZ:            elasticacheTypes.MultiAZStatusEnabled,
						Status:             aws.String("available"),
						MemberClusters:     []string{"prod-redis-001"},
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, &MockEKS{}, mockElastiCache, &MockELB{})

	resources, err := client.discoverElastiCacheClusters(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 ElastiCache cluster, got %d", len(resources))
	}

	resource := resources[0]
	if resource.ID != "prod-redis" {
		t.Errorf("expected ID prod-redis, got %s", resource.ID)
	}
	if resource.Type != "aws_elasticache_replication_group" {
		t.Errorf("expected Type aws_elasticache_replication_group, got %s", resource.Type)
	}
}

// TestDiscoverLoadBalancers_Success tests successful load balancer discovery
func TestDiscoverLoadBalancers_Success(t *testing.T) {
	mockELB := &MockELB{
		DescribeLoadBalancersFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
				LoadBalancers: []elbTypes.LoadBalancer{
					{
						LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/prod-lb/50dc6c495c0c9188"),
						LoadBalancerName: aws.String("prod-lb"),
						VpcId:            aws.String("vpc-12345"),
						DNSName:          aws.String("prod-lb-123456.us-east-1.elb.amazonaws.com"),
						State: &elbTypes.LoadBalancerState{
							Code: "active",
						},
						SecurityGroups: []string{"sg-12345"},
						AvailabilityZones: []elbTypes.AvailabilityZone{
							{
								ZoneName: aws.String("us-east-1a"),
								SubnetId: aws.String("subnet-12345"),
							},
						},
					},
				},
			}, nil
		},
		DescribeTagsFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
			return &elasticloadbalancingv2.DescribeTagsOutput{
				TagDescriptions: []elbTypes.TagDescription{
					{
						ResourceArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/prod-lb/50dc6c495c0c9188"),
						Tags: []elbTypes.Tag{
							{Key: aws.String("Name"), Value: aws.String("prod-lb")},
						},
					},
				},
			}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, mockELB)

	resources, err := client.discoverLoadBalancers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 load balancer, got %d", len(resources))
	}

	resource := resources[0]
	if resource.Type != "aws_lb" {
		t.Errorf("expected Type aws_lb, got %s", resource.Type)
	}
}

// TestDiscoverAll_Success tests successful discovery of all resources
func TestDiscoverAll_Success(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeVpcsFunc: func(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []ec2Types.Vpc{
					{VpcId: aws.String("vpc-12345"), CidrBlock: aws.String("10.0.0.0/16"), State: ec2Types.VpcStateAvailable},
				},
			}, nil
		},
		DescribeSubnetsFunc: func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
			return &ec2.DescribeSubnetsOutput{}, nil
		},
		DescribeSecurityGroupsFunc: func(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{}, nil
		},
		DescribeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{}, nil
		},
	}

	mockRDS := &MockRDS{
		DescribeDBInstancesFunc: func(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
			return &rds.DescribeDBInstancesOutput{}, nil
		},
	}

	mockEKS := &MockEKS{
		ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{}, nil
		},
	}

	mockElastiCache := &MockElastiCache{
		DescribeReplicationGroupsFunc: func(ctx context.Context, params *elasticache.DescribeReplicationGroupsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
			return &elasticache.DescribeReplicationGroupsOutput{}, nil
		},
	}

	mockELB := &MockELB{
		DescribeLoadBalancersFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, mockRDS, mockEKS, mockElastiCache, mockELB)

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) < 1 {
		t.Fatalf("expected at least 1 resource, got %d", len(resources))
	}
}

// TestDiscoverAll_PartialFailure tests DiscoverAll with some service failures
func TestDiscoverAll_PartialFailure(t *testing.T) {
	mockEC2 := &MockEC2{
		DescribeVpcsFunc: func(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []ec2Types.Vpc{
					{VpcId: aws.String("vpc-12345"), CidrBlock: aws.String("10.0.0.0/16"), State: ec2Types.VpcStateAvailable},
				},
			}, nil
		},
		DescribeSubnetsFunc: func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
			return nil, errors.New("access denied")
		},
		DescribeSecurityGroupsFunc: func(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{}, nil
		},
		DescribeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{}, nil
		},
	}

	mockRDS := &MockRDS{
		DescribeDBInstancesFunc: func(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
			return &rds.DescribeDBInstancesOutput{}, nil
		},
	}

	mockEKS := &MockEKS{
		ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{}, nil
		},
	}

	mockElastiCache := &MockElastiCache{
		DescribeReplicationGroupsFunc: func(ctx context.Context, params *elasticache.DescribeReplicationGroupsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
			return &elasticache.DescribeReplicationGroupsOutput{}, nil
		},
	}

	mockELB := &MockELB{
		DescribeLoadBalancersFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{}, nil
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", mockEC2, mockRDS, mockEKS, mockElastiCache, mockELB)

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on partial failure: %v", err)
	}

	// Should still return VPCs even though subnets failed
	if len(resources) < 1 {
		t.Fatalf("expected at least 1 resource despite partial failure, got %d", len(resources))
	}
}

// TestRDSError_DescribeDBInstances tests RDS error handling
func TestRDSError_DescribeDBInstances(t *testing.T) {
	mockRDS := &MockRDS{
		DescribeDBInstancesFunc: func(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
			return nil, errors.New("throttling error")
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, mockRDS, &MockEKS{}, &MockElastiCache{}, &MockELB{})

	_, err := client.discoverRDSInstances(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestEKSError_ListClusters tests EKS error handling on list
func TestEKSError_ListClusters(t *testing.T) {
	mockEKS := &MockEKS{
		ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return nil, errors.New("service error")
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, mockEKS, &MockElastiCache{}, &MockELB{})

	_, err := client.discoverEKSClusters(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestEKSError_DescribeCluster tests EKS error handling on describe
func TestEKSError_DescribeCluster(t *testing.T) {
	mockEKS := &MockEKS{
		ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"test-cluster"},
			}, nil
		},
		DescribeClusterFunc: func(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
			return nil, errors.New("cluster not found")
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, mockEKS, &MockElastiCache{}, &MockELB{})

	_, err := client.discoverEKSClusters(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestElastiCacheError tests ElastiCache error handling
func TestElastiCacheError(t *testing.T) {
	mockElastiCache := &MockElastiCache{
		DescribeReplicationGroupsFunc: func(ctx context.Context, params *elasticache.DescribeReplicationGroupsInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
			return nil, errors.New("describe error")
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, &MockEKS{}, mockElastiCache, &MockELB{})

	_, err := client.discoverElastiCacheClusters(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestLoadBalancerError tests ELB error handling
func TestLoadBalancerError(t *testing.T) {
	mockELB := &MockELB{
		DescribeLoadBalancersFunc: func(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return nil, errors.New("describe error")
		},
	}

	client := NewDiscoveryClientWithServices("us-east-1", &MockEC2{}, &MockRDS{}, &MockEKS{}, &MockElastiCache{}, mockELB)

	_, err := client.discoverLoadBalancers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestNewDiscoveryClientWithServices creates a discovery client with injected services
func TestNewDiscoveryClientWithServices(t *testing.T) {
	mockEC2 := &MockEC2{}
	mockRDS := &MockRDS{}
	mockEKS := &MockEKS{}
	mockElastiCache := &MockElastiCache{}
	mockELB := &MockELB{}

	client := NewDiscoveryClientWithServices("us-west-2", mockEC2, mockRDS, mockEKS, mockElastiCache, mockELB)

	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.region != "us-west-2" {
		t.Errorf("expected region us-west-2, got %s", client.region)
	}
}
