/**
 * Cytoscape Node Styles
 * AWS resource and general node styling
 */

import {
  AWS_NETWORK_COLORS,
  NODE_TYPE_COLORS,
  CYTOSCAPE_DRIFT_COLORS,
} from '../../constants/colors';

interface CytoscapeStyle {
  selector: string;
  style: Record<string, unknown>;
}

export const nodeStyles: CytoscapeStyle[] = [
  // Default node style
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
      'border-color': CYTOSCAPE_DRIFT_COLORS.low,
    },
  },

  // Terraform Resource
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
      'border-color': '#FF9900',
    },
  },

  // Terraform Change
  {
    selector: 'node[type="terraform_change"]',
    style: {
      'background-color': '#ffffff',
      'background-image':
        'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjNjIzQ0U0Ii8+PHBhdGggZD0iTTE2IDE0TDIwIDE2VjMyTDE2IDMwVjE0WiIgZmlsbD0id2hpdGUiLz48cGF0aCBkPSJNMjIgMTZMMjYgMThWMzRMMjIgMzJWMTZaIiBmaWxsPSJ3aGl0ZSIvPjxwYXRoIGQ9Ik0yOCAxNkwzMiAxOFYzNEwyOCAzMlYxNloiIGZpbGw9IndoaXRlIi8+PC9zdmc+',
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
      'border-color': NODE_TYPE_COLORS.terraformChange,
      'z-index': 100,
    },
  },

  // IAM Role
  {
    selector: 'node[type="iam_role"]',
    style: {
      'background-color': NODE_TYPE_COLORS.iamRole,
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 70,
      'height': 70,
      'font-size': '11px',
      'color': '#ffffff',
      'text-outline-color': '#1864ab',
      'text-outline-width': 2,
      'border-width': 2,
      'border-color': '#1864ab',
    },
  },

  // Service Account
  {
    selector: 'node[type="service_account"]',
    style: {
      'background-color': NODE_TYPE_COLORS.serviceAccount,
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 70,
      'height': 70,
      'font-size': '11px',
      'color': '#ffffff',
      'text-outline-color': '#2f9e44',
      'text-outline-width': 2,
      'border-width': 2,
      'border-color': '#2f9e44',
    },
  },

  // Pod
  {
    selector: 'node[type="pod"]',
    style: {
      'background-color': NODE_TYPE_COLORS.pod,
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 60,
      'height': 60,
      'font-size': '10px',
      'color': '#000000',
      'text-outline-color': '#fab005',
      'text-outline-width': 1,
      'border-width': 2,
      'border-color': CYTOSCAPE_DRIFT_COLORS.medium,
    },
  },

  // Container
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
      'border-color': CYTOSCAPE_DRIFT_COLORS.high,
    },
  },

  // Falco Event
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
      'z-index': 90,
    },
  },

  // VPC
  {
    selector: 'node[resource_type="aws_vpc"]',
    style: {
      'background-color': 'rgba(235, 245, 251, 0.95)',
      'background-image': '/aws-icons/vpc.svg',
      'background-fit': 'none',
      'background-image-opacity': 0.3,
      'background-position-x': '10px',
      'background-position-y': '10px',
      'background-width': '48px',
      'background-height': '48px',
      'border-color': AWS_NETWORK_COLORS.vpc,
      'border-width': 5,
      'border-style': 'dashed',
      'shape': 'roundrectangle',
      'padding': '100px',
      'text-valign': 'top',
      'text-halign': 'left',
      'text-margin-x': 10,
      'text-margin-y': 10,
      'color': AWS_NETWORK_COLORS.vpc,
      'font-size': '13px',
      'font-weight': 'bold',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.95,
      'text-background-padding': '3px',
      'text-background-shape': 'roundrectangle',
    },
  },

  // Subnet
  {
    selector: 'node[resource_type="aws_subnet"]',
    style: {
      'background-color': 'rgba(214, 234, 248, 0.9)',
      'background-image': '/aws-icons/subnet.svg',
      'background-fit': 'none',
      'background-image-opacity': 0.35,
      'background-position-x': '8px',
      'background-position-y': '8px',
      'background-width': '36px',
      'background-height': '36px',
      'border-color': AWS_NETWORK_COLORS.subnet,
      'border-width': 4,
      'border-style': 'solid',
      'shape': 'roundrectangle',
      'padding': '70px',
      'text-valign': 'top',
      'text-halign': 'left',
      'text-margin-x': 8,
      'text-margin-y': 8,
      'color': AWS_NETWORK_COLORS.vpc,
      'font-size': '12px',
      'font-weight': '600',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.95,
      'text-background-padding': '2px',
      'text-background-shape': 'roundrectangle',
    },
  },
];
