// +build tools

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// TerraformState represents a Terraform state file
type TerraformState struct {
	Version          int        `json:"version"`
	TerraformVersion string     `json:"terraform_version"`
	Serial           int        `json:"serial"`
	Lineage          string     `json:"lineage"`
	Outputs          struct{}   `json:"outputs"`
	Resources        []Resource `json:"resources"`
}

type Resource struct {
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Provider  string     `json:"provider"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	SchemaVersion int                    `json:"schema_version"`
	Attributes    map[string]interface{} `json:"attributes"`
}

var (
	resourceTypes = []ResourceType{
		// EC2 - 30%
		{Type: "aws_instance", Weight: 15},
		{Type: "aws_ebs_volume", Weight: 8},
		{Type: "aws_security_group", Weight: 7},

		// IAM - 20%
		{Type: "aws_iam_role", Weight: 8},
		{Type: "aws_iam_policy", Weight: 6},
		{Type: "aws_iam_user", Weight: 4},
		{Type: "aws_iam_role_policy_attachment", Weight: 2},

		// S3 - 15%
		{Type: "aws_s3_bucket", Weight: 10},
		{Type: "aws_s3_bucket_policy", Weight: 5},

		// RDS - 10%
		{Type: "aws_db_instance", Weight: 6},
		{Type: "aws_db_subnet_group", Weight: 4},

		// Lambda - 8%
		{Type: "aws_lambda_function", Weight: 8},

		// VPC - 10%
		{Type: "aws_vpc", Weight: 3},
		{Type: "aws_subnet", Weight: 4},
		{Type: "aws_route_table", Weight: 3},

		// Others - 7%
		{Type: "aws_lb", Weight: 3},
		{Type: "aws_lb_target_group", Weight: 2},
		{Type: "aws_cloudwatch_log_group", Weight: 2},
	}
)

type ResourceType struct {
	Type   string
	Weight int
}

func main() {
	count := flag.Int("resources", 500, "Number of resources to generate")
	output := flag.String("output", "terraform.tfstate", "Output file")
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	rand.Seed(*seed)

	fmt.Printf("Terraform State Generator\n")
	fmt.Printf("  Resources: %d\n", *count)
	fmt.Printf("  Output: %s\n", *output)
	fmt.Println()

	state := TerraformState{
		Version:          4,
		TerraformVersion: "1.5.0",
		Serial:           1,
		Lineage:          randomString(36),
		Resources:        make([]Resource, 0, *count),
	}

	fmt.Println("Generating resources...")

	for i := 0; i < *count; i++ {
		resourceType := selectResourceType()
		resource := generateResource(resourceType, i)
		state.Resources = append(state.Resources, resource)

		if (i+1)%1000 == 0 {
			fmt.Printf("  Generated %d/%d resources\n", i+1, *count)
		}
	}

	fmt.Printf("  Generated %d/%d resources\n", *count, *count)
	fmt.Println()

	// Write to file
	fmt.Println("Writing to file...")
	file, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write state: %v\n", err)
		os.Exit(1)
	}

	// Get file size
	info, _ := file.Stat()
	fmt.Printf("Done! File size: %.2f MB\n", float64(info.Size())/(1024*1024))
}

func selectResourceType() string {
	totalWeight := 0
	for _, rt := range resourceTypes {
		totalWeight += rt.Weight
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for _, rt := range resourceTypes {
		cumulative += rt.Weight
		if r < cumulative {
			return rt.Type
		}
	}

	return resourceTypes[0].Type
}

func generateResource(resourceType string, index int) Resource {
	name := fmt.Sprintf("%s-%d", resourceType, index)

	resource := Resource{
		Mode:      "managed",
		Type:      resourceType,
		Name:      name,
		Provider:  "provider[\"registry.terraform.io/hashicorp/aws\"]",
		Instances: []Instance{generateInstance(resourceType)},
	}

	return resource
}

func generateInstance(resourceType string) Instance {
	instance := Instance{
		SchemaVersion: 1,
		Attributes:    generateAttributes(resourceType),
	}
	return instance
}

func generateAttributes(resourceType string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Common attributes
	attrs["id"] = generateResourceID(resourceType)
	attrs["arn"] = generateARN(resourceType, attrs["id"].(string))
	attrs["tags"] = map[string]string{
		"Name":        fmt.Sprintf("%s-%s", resourceType, randomString(8)),
		"Environment": []string{"dev", "staging", "prod"}[rand.Intn(3)],
		"ManagedBy":   "terraform",
	}

	// Type-specific attributes
	switch resourceType {
	case "aws_instance":
		attrs["instance_type"] = []string{"t3.micro", "t3.small", "t3.medium", "m5.large"}[rand.Intn(4)]
		attrs["ami"] = fmt.Sprintf("ami-%s", randomString(17))
		attrs["disable_api_termination"] = rand.Intn(2) == 1
		attrs["vpc_security_group_ids"] = []string{fmt.Sprintf("sg-%s", randomString(17))}
		attrs["subnet_id"] = fmt.Sprintf("subnet-%s", randomString(17))

	case "aws_ebs_volume":
		attrs["size"] = rand.Intn(1000) + 10
		attrs["type"] = []string{"gp2", "gp3", "io1"}[rand.Intn(3)]
		attrs["encrypted"] = rand.Intn(2) == 1
		attrs["availability_zone"] = fmt.Sprintf("us-east-1%c", 'a'+rune(rand.Intn(6)))

	case "aws_security_group":
		attrs["name"] = fmt.Sprintf("sg-%s", randomString(12))
		attrs["description"] = "Managed by Terraform"
		attrs["vpc_id"] = fmt.Sprintf("vpc-%s", randomString(17))
		attrs["ingress"] = generateSecurityGroupRules(rand.Intn(5) + 1)
		attrs["egress"] = generateSecurityGroupRules(rand.Intn(3) + 1)

	case "aws_iam_role":
		attrs["name"] = fmt.Sprintf("role-%s", randomString(12))
		attrs["assume_role_policy"] = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"
		attrs["max_session_duration"] = 3600

	case "aws_iam_policy":
		attrs["name"] = fmt.Sprintf("policy-%s", randomString(12))
		attrs["policy"] = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

	case "aws_iam_user":
		attrs["name"] = fmt.Sprintf("user-%s", randomString(12))

	case "aws_s3_bucket":
		attrs["bucket"] = fmt.Sprintf("bucket-%s", randomString(16))
		attrs["versioning"] = map[string]interface{}{
			"enabled": rand.Intn(2) == 1,
		}
		attrs["server_side_encryption_configuration"] = map[string]interface{}{
			"rule": map[string]interface{}{
				"apply_server_side_encryption_by_default": map[string]string{
					"sse_algorithm": "AES256",
				},
			},
		}

	case "aws_db_instance":
		attrs["identifier"] = fmt.Sprintf("db-%s", randomString(12))
		attrs["engine"] = []string{"mysql", "postgres", "mariadb"}[rand.Intn(3)]
		attrs["instance_class"] = []string{"db.t3.micro", "db.t3.small", "db.m5.large"}[rand.Intn(3)]
		attrs["allocated_storage"] = rand.Intn(1000) + 20
		attrs["publicly_accessible"] = false

	case "aws_lambda_function":
		attrs["function_name"] = fmt.Sprintf("lambda-%s", randomString(12))
		attrs["runtime"] = []string{"python3.9", "nodejs18.x", "go1.x"}[rand.Intn(3)]
		attrs["handler"] = "index.handler"
		attrs["timeout"] = rand.Intn(900) + 3
		attrs["memory_size"] = []int{128, 256, 512, 1024, 2048}[rand.Intn(5)]
		attrs["role"] = generateARN("aws_iam_role", fmt.Sprintf("role-%s", randomString(12)))

	case "aws_vpc":
		attrs["cidr_block"] = fmt.Sprintf("10.%d.0.0/16", rand.Intn(256))
		attrs["enable_dns_support"] = true
		attrs["enable_dns_hostnames"] = true

	case "aws_subnet":
		attrs["cidr_block"] = fmt.Sprintf("10.%d.%d.0/24", rand.Intn(256), rand.Intn(256))
		attrs["vpc_id"] = fmt.Sprintf("vpc-%s", randomString(17))
		attrs["availability_zone"] = fmt.Sprintf("us-east-1%c", 'a'+rune(rand.Intn(6)))

	case "aws_route_table":
		attrs["vpc_id"] = fmt.Sprintf("vpc-%s", randomString(17))
		attrs["route"] = []map[string]interface{}{
			{
				"cidr_block": "0.0.0.0/0",
				"gateway_id": fmt.Sprintf("igw-%s", randomString(17)),
			},
		}

	case "aws_lb":
		attrs["name"] = fmt.Sprintf("lb-%s", randomString(12))
		attrs["load_balancer_type"] = []string{"application", "network"}[rand.Intn(2)]
		attrs["subnets"] = []string{
			fmt.Sprintf("subnet-%s", randomString(17)),
			fmt.Sprintf("subnet-%s", randomString(17)),
		}

	case "aws_lb_target_group":
		attrs["name"] = fmt.Sprintf("tg-%s", randomString(12))
		attrs["port"] = []int{80, 443, 8080}[rand.Intn(3)]
		attrs["protocol"] = "HTTP"
		attrs["vpc_id"] = fmt.Sprintf("vpc-%s", randomString(17))
	}

	return attrs
}

func generateResourceID(resourceType string) string {
	switch resourceType {
	case "aws_instance":
		return fmt.Sprintf("i-%s", randomString(17))
	case "aws_ebs_volume":
		return fmt.Sprintf("vol-%s", randomString(17))
	case "aws_security_group":
		return fmt.Sprintf("sg-%s", randomString(17))
	case "aws_iam_role", "aws_iam_policy", "aws_iam_user":
		return fmt.Sprintf("%s-%s", resourceType, randomString(12))
	case "aws_s3_bucket":
		return fmt.Sprintf("bucket-%s", randomString(16))
	case "aws_db_instance":
		return fmt.Sprintf("db-%s", randomString(12))
	case "aws_lambda_function":
		return fmt.Sprintf("lambda-%s", randomString(12))
	case "aws_vpc":
		return fmt.Sprintf("vpc-%s", randomString(17))
	case "aws_subnet":
		return fmt.Sprintf("subnet-%s", randomString(17))
	case "aws_route_table":
		return fmt.Sprintf("rtb-%s", randomString(17))
	case "aws_lb":
		return fmt.Sprintf("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/lb-%s/%s",
			randomString(12), randomString(16))
	case "aws_lb_target_group":
		return fmt.Sprintf("arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/tg-%s/%s",
			randomString(12), randomString(16))
	default:
		return randomString(20)
	}
}

func generateARN(resourceType, resourceID string) string {
	service := "iam"
	if resourceType == "aws_s3_bucket" {
		service = "s3"
		return fmt.Sprintf("arn:aws:s3:::%s", resourceID)
	} else if resourceType == "aws_lambda_function" {
		service = "lambda"
	} else if resourceType == "aws_db_instance" {
		service = "rds"
	}

	return fmt.Sprintf("arn:aws:%s::123456789012:%s/%s", service, resourceType, resourceID)
}

func generateSecurityGroupRules(count int) []map[string]interface{} {
	rules := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		rules[i] = map[string]interface{}{
			"from_port":   rand.Intn(65535),
			"to_port":     rand.Intn(65535),
			"protocol":    []string{"tcp", "udp", "icmp"}[rand.Intn(3)],
			"cidr_blocks": []string{"0.0.0.0/0"},
		}
	}
	return rules
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
