/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Cytoscape.js Style Definitions
 *
 * TFDrift-Falco因果関係グラフのビジュアルスタイル
 */

export const cytoscapeStylesheet: any[] = [
  // ==================== デフォルトスタイル ====================

  // Default node style - applies to all nodes
  // Architecture diagram sizing: optimized for visibility and clarity
  {
    selector: 'node',
    style: {
      'background-color': '#e9ecef',
      'label': 'data(label)',
      'shape': 'roundrectangle',
      'width': 60,
      'height': 60,
      'font-size': '11px',
      'text-valign': 'bottom',
      'text-halign': 'center',
      'text-margin-y': 6,
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '3px',
      'text-background-shape': 'roundrectangle',
      'border-width': 2,
      'border-color': '#adb5bd'
    }
  },

  // Default edge style - applies to all edges
  {
    selector: 'edge',
    style: {
      'width': 2,
      'line-color': '#adb5bd',
      'target-arrow-color': '#adb5bd',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': '10px',
      'text-rotation': 'autorotate',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.8,
      'text-background-padding': '3px'
    }
  },

  // ==================== ノードスタイル ====================

  // Generic Terraform Resource nodes (default for unmapped resources)
  // Changed to white background so undefined resources are visible
  {
    selector: 'node[type="terraform_resource"]',
    style: {
      'background-color': '#ffffff',
      'label': 'data(label)',
      'shape': 'roundrectangle',
      'width': 60,
      'height': 60,
      'font-size': '11px',
      'text-valign': 'bottom',
      'text-halign': 'center',
      'text-margin-y': 6,
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '3px',
      'text-background-shape': 'roundrectangle',
      'border-width': 2,
      'border-color': '#FF9900'
    }
  },

  // Terraform Change nodes - 起点（紫のアイコン、より大きい）
  {
    selector: 'node[type="terraform_change"]',
    style: {
      'background-color': '#ffffff',
      'background-image': 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjNjIzQ0U0Ii8+PHBhdGggZD0iTTE2IDE0TDIwIDE2VjMyTDE2IDMwVjE0WiIgZmlsbD0id2hpdGUiLz48cGF0aCBkPSJNMjIgMTZMMjYgMThWMzRMMjIgMzJWMTZaIiBmaWxsPSJ3aGl0ZSIvPjxwYXRoIGQ9Ik0yOCAxNkwzMiAxOFYzNEwyOCAzMlYxNloiIGZpbGw9IndoaXRlIi8+PC9zdmc+',
      'background-fit': 'contain',
      'background-clip': 'none',
      'label': 'data(label)',
      'shape': 'roundrectangle',
      'width': 90,
      'height': 90,
      'font-size': '11px',
      'text-valign': 'bottom',
      'text-halign': 'center',
      'text-margin-y': 8,
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '4px',
      'text-background-shape': 'roundrectangle',
      'border-width': 3,
      'border-color': '#623CE4',
      'z-index': 100
    }
  },

  // IAM Policy nodes (AWS IAMアイコン)
  {
    selector: 'node[type="iam_policy"]',
    style: {
      'background-color': '#ffffff',
      'background-image': 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjREQzNDRDIi8+PHBhdGggZD0iTTI0IDE2QzI3LjMxMzcgMTYgMzAgMTguNjg2MyAzMCAyMkMzMCAyNS4zMTM3IDI3LjMxMzcgMjggMjQgMjhDMjAuNjg2MyAyOCAxOCAyNS4zMTM3IDE4IDIyQzE4IDE4LjY4NjMgMjAuNjg2MyAxNiAyNCAxNloiIGZpbGw9IndoaXRlIi8+PHBhdGggZD0iTTMxIDMwSDE3QzE1Ljg5NTQgMzAgMTUgMzAuODk1NCAxNSAzMlYzM0MxNSAzMy41NTIzIDE1LjQ0NzcgMzQgMTYgMzRIMzJDMzIuNTUyMyAzNCAzMyAzMy41NTIzIDMzIDMzVjMyQzMzIDMwLjg5NTQgMzIuMTA0NiAzMCAzMSAzMFoiIGZpbGw9IndoaXRlIi8+PC9zdmc+',
      'background-fit': 'contain',
      'background-clip': 'none',
      'label': 'data(label)',
      'shape': 'roundrectangle',
      'width': 80,
      'height': 80,
      'font-size': '10px',
      'text-valign': 'bottom',
      'text-halign': 'center',
      'text-margin-y': 6,
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '3px',
      'text-background-shape': 'roundrectangle',
      'border-width': 2,
      'border-color': '#DD344C'
    }
  },

  // IAM Role nodes（濃い青）
  {
    selector: 'node[type="iam_role"]',
    style: {
      'background-color': '#1c7ed6',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 70,
      'height': 70,
      'font-size': '11px',
      'color': '#ffffff',
      'text-outline-color': '#1864ab',
      'text-outline-width': 2,
      'border-width': 2,
      'border-color': '#1864ab'
    }
  },

  // Service Account nodes（緑の角丸四角）
  {
    selector: 'node[type="service_account"]',
    style: {
      'background-color': '#51cf66',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 70,
      'height': 70,
      'font-size': '11px',
      'color': '#ffffff',
      'text-outline-color': '#2f9e44',
      'text-outline-width': 2,
      'border-width': 2,
      'border-color': '#2f9e44'
    }
  },

  // Pod nodes（黄色の角丸四角）
  {
    selector: 'node[type="pod"]',
    style: {
      'background-color': '#ffd43b',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 60,
      'height': 60,
      'font-size': '10px',
      'color': '#000000',
      'text-outline-color': '#fab005',
      'text-outline-width': 1,
      'border-width': 2,
      'border-color': '#fab005'
    }
  },

  // Container nodes（オレンジの楕円）
  {
    selector: 'node[type="container"]',
    style: {
      'background-color': '#ff922b',
      'label': 'data(label)',
      'shape': 'ellipse',
      'width': 55,
      'height': 55,
      'font-size': '11px',
      'color': '#ffffff',
      'text-outline-color': '#e8590c',
      'text-outline-width': 1,
      'border-width': 2,
      'border-color': '#e8590c'
    }
  },

  // Falco Event nodes（ピンクのダイアモンド）
  {
    selector: 'node[type="falco_event"]',
    style: {
      'background-color': '#f06595',
      'label': 'data(label)',
      'shape': 'diamond',
      'width': 70,
      'height': 70,
      'font-size': '10px',
      'color': '#ffffff',
      'text-outline-color': '#c2255c',
      'text-outline-width': 2,
      'border-width': 3,
      'border-color': '#c2255c',
      'z-index': 90
    }
  },

  // Security Group nodes（紫の角丸四角）
  {
    selector: 'node[type="security_group"]',
    style: {
      'background-color': '#9775fa',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 65,
      'height': 65,
      'font-size': '10px',
      'color': '#ffffff',
      'text-outline-color': '#6741d9',
      'text-outline-width': 2,
      'border-width': 2,
      'border-color': '#6741d9'
    }
  },

  // Network nodes（水色の楕円）
  {
    selector: 'node[type="network"]',
    style: {
      'background-color': '#22b8cf',
      'label': 'data(label)',
      'shape': 'ellipse',
      'width': 65,
      'height': 65,
      'font-size': '10px',
      'color': '#ffffff',
      'text-outline-color': '#0b7285',
      'text-outline-width': 2,
      'border-width': 2,
      'border-color': '#0b7285'
    }
  },

  // ==================== AWS Service Specific Styles ====================

  // VPC - AWS VPC Blue with official icon (Compound node - container)
  // Architecture diagram: Large container, prominent in hierarchy
  {
    selector: 'node[resource_type="aws_vpc"]',
    style: {
      'background-color': 'rgba(235, 245, 251, 0.95)',  // Very high opacity for maximum visibility
      'background-image': '/aws-icons/vpc.svg',
      'background-fit': 'none',
      'background-image-opacity': 0.3,
      'background-position-x': '10px',
      'background-position-y': '10px',
      'background-width': '48px',
      'background-height': '48px',
      'border-color': '#2E73B8',
      'border-width': 5,  // Thicker border for large graphs
      'border-style': 'dashed',  // Dashed border for VPC
      'shape': 'roundrectangle',
      'padding': '100px',  // Even larger padding for large graphs
      'text-valign': 'top',
      'text-halign': 'left',
      'text-margin-x': 10,
      'text-margin-y': 10,
      'color': '#2E73B8',
      'font-size': '13px',
      'font-weight': 'bold',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.95,
      'text-background-padding': '3px',
      'text-background-shape': 'roundrectangle'
    }
  },

  // Subnet - AWS VPC Blue (lighter) with official icon (Compound node - container)
  // Architecture diagram: Medium container within VPC
  {
    selector: 'node[resource_type="aws_subnet"]',
    style: {
      'background-color': 'rgba(214, 234, 248, 0.9)',  // Very high opacity for clear visibility
      'background-image': '/aws-icons/subnet.svg',
      'background-fit': 'none',
      'background-image-opacity': 0.35,
      'background-position-x': '8px',
      'background-position-y': '8px',
      'background-width': '36px',
      'background-height': '36px',
      'border-color': '#5294CF',
      'border-width': 4,  // Thicker border
      'border-style': 'solid',  // Solid border for Subnet
      'shape': 'roundrectangle',
      'padding': '70px',  // Larger padding for large graphs
      'text-valign': 'top',
      'text-halign': 'left',
      'text-margin-x': 8,
      'text-margin-y': 8,
      'color': '#2E73B8',
      'font-size': '12px',
      'font-weight': '600',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.95,
      'text-background-padding': '2px',
      'text-background-shape': 'roundrectangle'
    }
  },

  // Security Group - AWS Security Red with official icon
  {
    selector: 'node[resource_type="aws_security_group"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/security-group.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#DD344C',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'border-width': 3,
      'font-size': '11px',
      'font-weight': 'bold'
    }
  },

  // EKS Cluster - AWS Container Orange with official icon (Large infrastructure)
  {
    selector: 'node[resource_type="aws_eks_cluster"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/eks.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#ED7100',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3,
      'font-weight': 'bold'
    }
  },

  // EKS Node Group - AWS Container Orange (similar to cluster) (Medium)
  {
    selector: 'node[resource_type="aws_eks_node_group"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/eks.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#ED7100',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3,
      'font-weight': 'bold'
    }
  },

  // EKS Addon - AWS Container Orange (Small utility for addons like CoreDNS)
  {
    selector: 'node[resource_type="aws_eks_addon"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/eks.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#ED7100',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // ECS - AWS Container Orange (lighter) with official icon (Large infrastructure)
  {
    selector: 'node[resource_type="aws_ecs_cluster"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/ecs.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#FF9900',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3,
      'font-weight': 'bold'
    }
  },

  // RDS - AWS Database Blue with official icon (Medium)
  {
    selector: 'node[resource_type="aws_db_instance"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/rds.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#3B48CC',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3,
      'font-weight': 'bold'
    }
  },

  // ElastiCache - AWS Database Blue (lighter) with official icon (Medium)
  {
    selector: 'node[resource_type="aws_elasticache_replication_group"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/elasticache.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#5A6EDB',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // Load Balancer - AWS Network Purple with official icon (Large infrastructure)
  {
    selector: 'node[resource_type="aws_lb"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/elb.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#8C4FFF',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3,
      'font-weight': 'bold'
    }
  },

  // Internet Gateway - AWS Network Green with official icon (Medium)
  {
    selector: 'node[resource_type="aws_internet_gateway"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/internet-gateway.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#7AA116',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // NAT Gateway - AWS Network Yellow-Green with official icon (Medium)
  {
    selector: 'node[resource_type="aws_nat_gateway"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/nat-gateway.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#B7CA18',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // Route Table - AWS Network Light Blue with official icon (Small utility)
  {
    selector: 'node[resource_type="aws_route_table"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/route-table.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#4A90E2',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // Route (individual route in route table) - Same icon as route table
  {
    selector: 'node[resource_type="aws_route"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/route-table.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#4A90E2',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // IAM Role - AWS Identity Red-Pink with official icon (Medium)
  {
    selector: 'node[resource_type="aws_iam_role"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/iam.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#DD344C',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3,
      'font-weight': 'bold'
    }
  },

  // IAM Policy - AWS Identity Pink with official icon (Small utility)
  {
    selector: 'node[resource_type="aws_iam_policy"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/iam.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#E94B8B',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // CloudWatch - AWS Management Orange with official icon (Small utility)
  {
    selector: 'node[resource_type="aws_cloudwatch_log_group"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/cloudwatch.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#E97623',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // KMS Key - AWS Security Green with official icon (Small utility)
  {
    selector: 'node[resource_type="aws_kms_key"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/kms.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#759C3E',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // KMS Alias - AWS Security Green with official icon (Small utility)
  {
    selector: 'node[resource_type="aws_kms_alias"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/kms.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#759C3E',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // Secrets Manager - AWS Security Green (lighter) with official icon (Small utility)
  {
    selector: 'node[resource_type="aws_secretsmanager_secret"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/secrets-manager.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#87B84E',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // ==================== 追加サービス ====================

  // Lambda - AWS Compute Orange with official icon (Medium)
  {
    selector: 'node[resource_type="aws_lambda_function"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/lambda.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#FF9900',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // S3 Bucket - AWS Storage Green with official icon (Medium)
  {
    selector: 'node[resource_type="aws_s3_bucket"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/s3.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#569A31',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // DynamoDB - AWS Database Blue with official icon (Medium)
  {
    selector: 'node[resource_type="aws_dynamodb_table"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/dynamodb.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#3B48CC',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // SNS Topic - AWS Integration Pink with official icon (Small)
  {
    selector: 'node[resource_type="aws_sns_topic"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/sns.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#D13212',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // SQS Queue - AWS Integration Pink with official icon (Small)
  {
    selector: 'node[resource_type="aws_sqs_queue"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/sqs.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#D13212',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // API Gateway - AWS Integration Purple with official icon (Medium)
  {
    selector: 'node[resource_type="aws_api_gateway_rest_api"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/api-gateway.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#945DF2',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // CloudFront Distribution - AWS Network Purple with official icon (Medium)
  {
    selector: 'node[resource_type="aws_cloudfront_distribution"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/cloudfront.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#8C4FFF',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // Step Functions - AWS Workflow Pink with official icon (Medium)
  {
    selector: 'node[resource_type="aws_sfn_state_machine"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/step-functions.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#D13212',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // EventBridge Rule - AWS Integration Pink with official icon (Small)
  {
    selector: 'node[resource_type="aws_cloudwatch_event_rule"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/eventbridge.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#D13212',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 70,
      'height': 70,
      'shape': 'roundrectangle',
      'font-size': '10px',
      'border-width': 2
    }
  },

  // Aurora Cluster - AWS Database Blue with official icon (Medium)
  {
    selector: 'node[resource_type="aws_rds_cluster"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/aurora.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#3B48CC',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // Fargate - AWS Compute Orange with official icon (Medium)
  {
    selector: 'node[resource_type="aws_ecs_service"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/fargate.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#FF9900',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // Neptune - AWS Database Blue with official icon (Medium)
  {
    selector: 'node[resource_type="aws_neptune_cluster"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/neptune.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#3B48CC',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // Timestream - AWS Database Blue with official icon (Medium)
  {
    selector: 'node[resource_type="aws_timestreamwrite_table"]',
    style: {
      'background-color': '#ffffff',
      'background-image': '/aws-icons/timestream.svg',
      'background-fit': 'contain',
      'background-clip': 'none',
      'background-width': '80%',
      'background-height': '80%',
      'border-color': '#3B48CC',
      'color': '#2d3748',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '2px',
      'width': 65,
      'height': 65,
      'shape': 'roundrectangle',
      'font-size': '11px',
      'border-width': 3
    }
  },

  // ==================== 重要度スタイル ====================

  // Critical severity - 強調ボーダー
  {
    selector: 'node[severity="critical"]',
    style: {
      'border-width': 6,
      'border-color': '#c92a2a'
    }
  },

  // High severity
  {
    selector: 'node[severity="high"]',
    style: {
      'border-width': 4,
      'border-color': '#e8590c'
    }
  },

  // Medium severity
  {
    selector: 'node[severity="medium"]',
    style: {
      'border-width': 2,
      'border-color': '#fab005'
    }
  },

  // Low severity
  {
    selector: 'node[severity="low"]',
    style: {
      'border-width': 1,
      'border-color': '#adb5bd'
    }
  },

  // ==================== エッジスタイル ====================

  // caused_by - Drift → IAM (太い赤)
  {
    selector: 'edge[type="caused_by"]',
    style: {
      'width': 4,
      'line-color': '#ff6b6b',
      'target-arrow-color': '#ff6b6b',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': '11px',
      'text-rotation': 'autorotate',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.8,
      'text-background-padding': '3px'
    }
  },

  // grants_access - IAM → ServiceAccount (青)
  {
    selector: 'edge[type="grants_access"]',
    style: {
      'width': 3,
      'line-color': '#4dabf7',
      'target-arrow-color': '#4dabf7',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': '10px',
      'text-rotation': 'autorotate',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.8,
      'text-background-padding': '3px'
    }
  },

  // used_by - SA → Pod (緑)
  {
    selector: 'edge[type="used_by"]',
    style: {
      'width': 3,
      'line-color': '#51cf66',
      'target-arrow-color': '#51cf66',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': '10px',
      'text-rotation': 'autorotate',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.8,
      'text-background-padding': '3px'
    }
  },

  // contains - Pod → Container (オレンジ)
  {
    selector: 'edge[type="contains"]',
    style: {
      'width': 2,
      'line-color': '#ff922b',
      'target-arrow-color': '#ff922b',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': '10px',
      'text-rotation': 'autorotate'
    }
  },

  // triggered - Container → Falco Event (ピンク、太い)
  {
    selector: 'edge[type="triggered"]',
    style: {
      'width': 4,
      'line-color': '#f06595',
      'target-arrow-color': '#f06595',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': '11px',
      'text-rotation': 'autorotate',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.8,
      'text-background-padding': '3px'
    }
  },

  // ==================== インタラクション状態 ====================

  // Highlighted path (パス強調)
  {
    selector: '.highlighted',
    style: {
      'background-color': '#ffd43b',
      'line-color': '#ffd43b',
      'target-arrow-color': '#ffd43b',
      'width': 6,
      'z-index': 999,
      'overlay-opacity': 0.3,
      'overlay-color': '#ffd43b',
      'overlay-padding': 5
    }
  },

  // Selected node (選択中のノード)
  {
    selector: ':selected',
    style: {
      'border-width': 6,
      'border-color': '#228be6',
      'z-index': 999,
      'overlay-opacity': 0.2,
      'overlay-color': '#228be6',
      'overlay-padding': 8
    }
  },

  // Hovered node (ホバー時)
  {
    selector: 'node:active',
    style: {
      'overlay-opacity': 0.2,
      'overlay-color': '#495057',
      'overlay-padding': 6
    }
  },

  // Dimmed nodes (非アクティブ)
  {
    selector: '.dimmed',
    style: {
      'opacity': 0.3
    }
  },

  // Blast radius center (Blast Radius中心)
  {
    selector: '.blast-center',
    style: {
      'background-color': '#fa5252',
      'border-width': 8,
      'border-color': '#c92a2a',
      'z-index': 1000
    }
  },

  // Blast radius affected (影響を受けるリソース)
  {
    selector: '.blast-affected',
    style: {
      'border-color': '#fa5252',
      'border-width': 4,
      'border-style': 'dashed'
    }
  },

  // ==================== Compound Node Overrides ====================

  // Ensure parent nodes (VPC/Subnet) are styled as compound containers
  {
    selector: ':parent',
    style: {
      'text-valign': 'top',
      'text-halign': 'left',
      'shape': 'roundrectangle',
      'compound-sizing-wrt-labels': 'include'
      // No min-width/min-height - let Cytoscape auto-size based on children
    }
  },

  // Ensure child nodes remain visible
  {
    selector: ':child',
    style: {
      'z-index': 999
    }
  }
];

// レイアウト設定
export const layoutConfigs = {
  // fCoSE - Best for compound nodes (VPC/Subnet hierarchy)
  // Optimized for architecture diagram with small nodes (40-50px)
  fcose: {
    name: 'fcose',
    quality: 'default',
    randomize: false,
    animate: true,
    animationDuration: 1000,
    animationEasing: 'ease-out',
    fit: true,
    padding: 50,
    nodeSeparation: 100,  // Increased for large graphs (100+ nodes)
    idealEdgeLength: 120,  // Increased for better spacing
    edgeElasticity: 0.45,
    nestingFactor: 0.2,  // Increased for better compound node spacing
    gravity: 0.25,  // Slightly reduced for larger graphs
    numIter: 3000,  // More iterations for large graphs
    tile: true,
    tilingPaddingVertical: 15,  // Increased padding
    tilingPaddingHorizontal: 15,
    gravityRangeCompound: 2.0,  // Increased for compound node cohesion
    gravityCompound: 1.5,  // Increased to keep children inside parents
    gravityRange: 4.0,
    initialEnergyOnIncremental: 0.3
  },

  dagre: {
    name: 'dagre',
    rankDir: 'TB',  // Top to Bottom
    nodeSep: 120,   // Increased spacing for better visibility
    rankSep: 180,   // Increased vertical spacing
    animate: true,
    animationDuration: 500,
    animationEasing: 'ease-out',
    // Compound node settings - critical for VPC/Subnet hierarchy
    nestingFactor: 1.2,  // How much extra space compound nodes get
    edgeSep: 20,
    ranker: 'longest-path',
    // Ensure compound nodes are respected
    fit: true,
    padding: 50
  },

  concentric: {
    name: 'concentric',
    concentric: (node: any) => node.data('distance') || 1,
    levelWidth: () => 2,
    animate: true,
    animationDuration: 500,
    minNodeSpacing: 100
  },

  cose: {
    name: 'cose',
    nodeRepulsion: 8000,
    idealEdgeLength: 100,
    animate: true,
    animationDuration: 1000,
    nodeOverlap: 20,
    gravity: 0.1
  },

  grid: {
    name: 'grid',
    rows: undefined,
    cols: undefined,
    animate: true,
    animationDuration: 500
  }
};

// Cytoscape.jsのコア設定
export const cytoscapeConfig = {
  style: cytoscapeStylesheet,
  minZoom: 0.1,
  maxZoom: 3,
  boxSelectionEnabled: true,
  autounselectify: false,
  autoungrabify: false,
  // Enable compound nodes for VPC/Subnet hierarchy
  compound: true
};
