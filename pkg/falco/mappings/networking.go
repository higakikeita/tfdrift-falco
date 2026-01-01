package mappings

// NetworkingMappings contains CloudTrail event to Terraform resource mappings for networking services
var NetworkingMappings = map[string]string{
	// VPC - Security Groups
	"AuthorizeSecurityGroupIngress": "aws_security_group",
	"AuthorizeSecurityGroupEgress":  "aws_security_group",
	"RevokeSecurityGroupIngress":    "aws_security_group",
	"RevokeSecurityGroupEgress":     "aws_security_group",
	"CreateSecurityGroup":           "aws_security_group",
	"DeleteSecurityGroup":           "aws_security_group",
	"ModifySecurityGroupRules":      "aws_security_group_rule",

	// VPC - Core
	"CreateVpc":             "aws_vpc",
	"DeleteVpc":             "aws_vpc",
	"ModifyVpcAttribute":    "aws_vpc",
	"CreateSubnet":          "aws_subnet",
	"DeleteSubnet":          "aws_subnet",
	"ModifySubnetAttribute": "aws_subnet",

	// VPC - Route Tables
	"CreateRoute":         "aws_route",
	"DeleteRoute":         "aws_route",
	"ReplaceRoute":        "aws_route",
	"CreateRouteTable":    "aws_route_table",
	"DeleteRouteTable":    "aws_route_table",
	"AssociateRouteTable": "aws_route_table_association",

	// VPC - Internet/NAT Gateways
	"AttachInternetGateway": "aws_internet_gateway_attachment",
	"DetachInternetGateway": "aws_internet_gateway_attachment",
	"CreateNatGateway":      "aws_nat_gateway",
	"DeleteNatGateway":      "aws_nat_gateway",
	"CreateInternetGateway": "aws_internet_gateway",
	"DeleteInternetGateway": "aws_internet_gateway",

	// VPC - Network ACLs
	"CreateNetworkAcl":       "aws_network_acl",
	"DeleteNetworkAcl":       "aws_network_acl",
	"CreateNetworkAclEntry":  "aws_network_acl_rule",
	"DeleteNetworkAclEntry":  "aws_network_acl_rule",
	"ReplaceNetworkAclEntry": "aws_network_acl_rule",

	// VPC - VPC Endpoints
	"CreateVpcEndpoint": "aws_vpc_endpoint",
	"DeleteVpcEndpoint": "aws_vpc_endpoint",
	"ModifyVpcEndpoint": "aws_vpc_endpoint",

	// VPC - VPC Peering
	"CreateVpcPeeringConnection": "aws_vpc_peering_connection",
	"DeleteVpcPeeringConnection": "aws_vpc_peering_connection",
	"AcceptVpcPeeringConnection": "aws_vpc_peering_connection_accepter",
	"RejectVpcPeeringConnection": "aws_vpc_peering_connection",

	// VPC - Transit Gateway
	"CreateTransitGateway":              "aws_ec2_transit_gateway",
	"DeleteTransitGateway":              "aws_ec2_transit_gateway",
	"ModifyTransitGateway":              "aws_ec2_transit_gateway",
	"CreateTransitGatewayVpcAttachment": "aws_ec2_transit_gateway_vpc_attachment",
	"DeleteTransitGatewayVpcAttachment": "aws_ec2_transit_gateway_vpc_attachment",
	"CreateTransitGatewayRouteTable":    "aws_ec2_transit_gateway_route_table",
	"DeleteTransitGatewayRouteTable":    "aws_ec2_transit_gateway_route_table",
	"CreateTransitGatewayRoute":         "aws_ec2_transit_gateway_route",
	"DeleteTransitGatewayRoute":         "aws_ec2_transit_gateway_route",

	// VPC - Flow Logs
	"CreateFlowLogs": "aws_flow_log",
	"DeleteFlowLogs": "aws_flow_log",

	// VPC - Network Firewall
	"CreateFirewall":       "aws_networkfirewall_firewall",
	"DeleteFirewall":       "aws_networkfirewall_firewall",
	"UpdateFirewall":       "aws_networkfirewall_firewall",
	"CreateFirewallPolicy": "aws_networkfirewall_firewall_policy",
	"DeleteFirewallPolicy": "aws_networkfirewall_firewall_policy",
	"UpdateFirewallPolicy": "aws_networkfirewall_firewall_policy",

	// ELB/ALB - Load Balancers
	"CreateLoadBalancer":           "aws_lb",
	"DeleteLoadBalancer":           "aws_lb",
	"ModifyLoadBalancerAttributes": "aws_lb",

	// ELB/ALB - Target Groups
	"CreateTargetGroup":           "aws_lb_target_group",
	"DeleteTargetGroup":           "aws_lb_target_group",
	"ModifyTargetGroup":           "aws_lb_target_group",
	"ModifyTargetGroupAttributes": "aws_lb_target_group",
	"RegisterTargets":             "aws_lb_target_group_attachment",
	"DeregisterTargets":           "aws_lb_target_group_attachment",

	// ELB/ALB - Listeners & Rules
	"CreateListener": "aws_lb_listener",
	"DeleteListener": "aws_lb_listener",
	"ModifyListener": "aws_lb_listener",
	"CreateRule":     "aws_lb_listener_rule",
	"ModifyRule":     "aws_lb_listener_rule",

	// ELB/ALB - Certificates
	"AddListenerCertificates":    "aws_lb_listener_certificate",
	"RemoveListenerCertificates": "aws_lb_listener_certificate",

	// Route53 - Hosted Zones
	"CreateHostedZone": "aws_route53_zone",
	"DeleteHostedZone": "aws_route53_zone",

	// Route53 - Record Sets
	"ChangeResourceRecordSets": "aws_route53_record",

	// Route53 - VPC Associations
	"AssociateVPCWithHostedZone":    "aws_route53_zone_association",
	"DisassociateVPCFromHostedZone": "aws_route53_zone_association",

	// Route53 - Tags
	"ChangeTagsForResource": "aws_route53_zone",

	// CloudFront - Distributions
	"CreateDistribution": "aws_cloudfront_distribution",
	"DeleteDistribution": "aws_cloudfront_distribution",
	"UpdateDistribution": "aws_cloudfront_distribution",

	// CloudFront - Invalidations
	"CreateInvalidation": "aws_cloudfront_invalidation",
}
