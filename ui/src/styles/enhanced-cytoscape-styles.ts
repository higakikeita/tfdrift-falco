/**
 * Enhanced Cytoscape Styles with Cloud Provider Icons
 *
 * Professional graph visualization with AWS/GCP/K8s official icons
 */

// Icon Data URLs (Base64 encoded SVGs)
const ICONS = {
  terraform: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjNjIzQ0U0Ii8+PHBhdGggZD0iTTE2IDE0TDIwIDE2VjMyTDE2IDMwVjE0WiIgZmlsbD0id2hpdGUiLz48cGF0aCBkPSJNMjIgMTZMMjYgMThWMzRMMjIgMzJWMTZaIiBmaWxsPSJ3aGl0ZSIvPjxwYXRoIGQ9Ik0yOCAxNkwzMiAxOFYzNEwyOCAzMlYxNloiIGZpbGw9IndoaXRlIi8+PC9zdmc+',

  aws_iam: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjREQzNDRDIi8+PHBhdGggZD0iTTI0IDE2QzI3LjMxMzcgMTYgMzAgMTguNjg2MyAzMCAyMkMzMCAyNS4zMTM3IDI3LjMxMzcgMjggMjQgMjhDMjAuNjg2MyAyOCAxOCAyNS4zMTM3IDE4IDIyQzE4IDE4LjY4NjMgMjAuNjg2MyAxNiAyNCAxNloiIGZpbGw9IndoaXRlIi8+PHBhdGggZD0iTTMxIDMwSDE3QzE1Ljg5NTQgMzAgMTUgMzAuODk1NCAxNSAzMlYzM0MxNSAzMy41NTIzIDE1LjQ0NzcgMzQgMTYgMzRIMzJDMzIuNTUyMyAzNCAzMyAzMy41NTIzIDMzIDMzVjMyQzMzIDMwLjg5NTQgMzIuMTA0NiAzMCAzMSAzMFoiIGZpbGw9IndoaXRlIi8+PC9zdmc+',

  kubernetes: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjMzI2Q0U1Ii8+PGNpcmNsZSBjeD0iMjQiIGN5PSIyNCIgcj0iMTIiIHN0cm9rZT0id2hpdGUiIHN0cm9rZS13aWR0aD0iMiIgZmlsbD0ibm9uZSIvPjxwYXRoIGQ9Ik0yNCAxNEwyOCAyNEwyNCAzNEwyMCAyNEwyNCAxNFoiIGZpbGw9IndoaXRlIi8+PGNpcmNsZSBjeD0iMjQiIGN5PSIyNCIgcj0iMyIgZmlsbD0id2hpdGUiLz48L3N2Zz4=',

  container: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjMERCN0VEIi8+PHJlY3QgeD0iMTIiIHk9IjE2IiB3aWR0aD0iMjQiIGhlaWdodD0iMTYiIHJ4PSIyIiBzdHJva2U9IndoaXRlIiBzdHJva2Utd2lkdGg9IjIiIGZpbGw9Im5vbmUiLz48cGF0aCBkPSJNMTIgMjJIMzYiIHN0cm9rZT0id2hpdGUiIHN0cm9rZS13aWR0aD0iMiIvPjxwYXRoIGQ9Ik0xMiAyNkgzNiIgc3Ryb2tlPSJ3aGl0ZSIgc3Ryb2tlLXdpZHRoPSIyIi8+PGNpcmNsZSBjeD0iMTYiIGN5PSIxOSIgcj0iMSIgZmlsbD0id2hpdGUiLz48Y2lyY2xlIGN4PSIyMCIgY3k9IjE5IiByPSIxIiBmaWxsPSJ3aGl0ZSIvPjxjaXJjbGUgY3g9IjI0IiBjeT0iMTkiIHI9IjEiIGZpbGw9IndoaXRlIi8+PC9zdmc+',

  falco: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI0IiBmaWxsPSIjRUM0QzM5Ii8+PGNpcmNsZSBjeD0iMjQiIGN5PSIyNCIgcj0iMTQiIHN0cm9rZT0id2hpdGUiIHN0cm9rZS13aWR0aD0iMiIgZmlsbD0ibm9uZSIvPjxwYXRoIGQ9Ik0yNCAxNkwyOCAyNEwyNCAzMkwyMCAyNEwyNCAxNloiIGZpbGw9IndoaXRlIi8+PGNpcmNsZSBjeD0iMjQiIGN5PSIyNCIgcj0iMyIgZmlsbD0iI0VDNEMzOSIvPjxwYXRoIGQ9Ik0yNCAxNlYxMk0yNCAzNlYzMk0xNiAyNEgxMk0zNiAyNEgzMiIgc3Ryb2tlPSJ3aGl0ZSIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiLz48L3N2Zz4='
};

// Base node style with modern design
const baseNodeStyle = {
  'background-color': '#ffffff',
  'background-fit': 'contain',
  'background-clip': 'none',
  'label': 'data(label)',
  'shape': 'roundrectangle',
  'font-family': '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto',
  'font-size': '10px',
  'font-weight': '600',
  'text-valign': 'bottom',
  'text-halign': 'center',
  'text-margin-y': 6,
  'color': '#1a202c',
  'text-background-color': '#ffffff',
  'text-background-opacity': 0.95,
  'text-background-padding': '4px',
  'text-background-shape': 'roundrectangle',
  'text-wrap': 'wrap',
  'text-max-width': '100px',
  'border-width': 2,
  'transition-property': 'border-width, border-color, width, height',
  'transition-duration': '0.2s'
};

export const enhancedCytoscapeStylesheet: any[] = [
  // ====================ノードスタイル (アイコンベース) ====================

  // Terraform Change
  {
    selector: 'node[type="terraform_change"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.terraform,
      'width': 90,
      'height': 90,
      'border-color': '#623CE4',
      'border-width': 3
    }
  },

  // IAM Policy & Role
  {
    selector: 'node[type="iam_policy"], node[type="iam_role"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.aws_iam,
      'width': 80,
      'height': 80,
      'border-color': '#DD344C'
    }
  },

  // Service Account
  {
    selector: 'node[type="service_account"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.kubernetes,
      'width': 75,
      'height': 75,
      'border-color': '#326CE5'
    }
  },

  // Pod
  {
    selector: 'node[type="pod"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.kubernetes,
      'width': 70,
      'height': 70,
      'border-color': '#326CE5'
    }
  },

  // Container
  {
    selector: 'node[type="container"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.container,
      'width': 65,
      'height': 65,
      'border-color': '#0DB7ED'
    }
  },

  // Falco Event
  {
    selector: 'node[type="falco_event"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.falco,
      'width': 85,
      'height': 85,
      'border-color': '#EC4C39',
      'border-width': 3,
      'z-index': 90
    }
  },

  // Security Group
  {
    selector: 'node[type="security_group"]',
    style: {
      ...baseNodeStyle,
      'background-image': ICONS.aws_iam,
      'width': 70,
      'height': 70,
      'border-color': '#9C27B0'
    }
  },

  // ==================== 重要度スタイル ====================

  {
    selector: 'node[severity="critical"]',
    style: {
      'border-width': 5,
      'border-color': '#c92a2a'
    }
  },

  {
    selector: 'node[severity="high"]',
    style: {
      'border-width': 4,
      'border-color': '#e8590c'
    }
  },

  // ==================== エッジスタイル (モダンデザイン) ====================

  {
    selector: 'edge',
    style: {
      'width': 3,
      'curve-style': 'bezier',
      'target-arrow-shape': 'triangle',
      'arrow-scale': 1.2,
      'label': 'data(label)',
      'font-size': '9px',
      'font-weight': '500',
      'text-rotation': 'autorotate',
      'text-background-color': '#ffffff',
      'text-background-opacity': 0.9,
      'text-background-padding': '3px',
      'text-background-shape': 'roundrectangle',
      'line-style': 'solid',
      'transition-property': 'width, line-color',
      'transition-duration': '0.2s'
    }
  },

  {
    selector: 'edge[type="caused_by"]',
    style: {
      'line-color': '#ff6b6b',
      'target-arrow-color': '#ff6b6b',
      'width': 4
    }
  },

  {
    selector: 'edge[type="grants_access"]',
    style: {
      'line-color': '#4dabf7',
      'target-arrow-color': '#4dabf7'
    }
  },

  {
    selector: 'edge[type="used_by"]',
    style: {
      'line-color': '#51cf66',
      'target-arrow-color': '#51cf66'
    }
  },

  {
    selector: 'edge[type="contains"]',
    style: {
      'line-color': '#ff922b',
      'target-arrow-color': '#ff922b',
      'width': 2
    }
  },

  {
    selector: 'edge[type="triggered"]',
    style: {
      'line-color': '#f06595',
      'target-arrow-color': '#f06595',
      'width': 4
    }
  },

  // ==================== インタラクション状態 ====================

  {
    selector: '.highlighted',
    style: {
      'background-color': '#fef3c7',
      'border-color': '#f59e0b',
      'border-width': 5,
      'line-color': '#f59e0b',
      'target-arrow-color': '#f59e0b',
      'width': 6,
      'z-index': 999
    }
  },

  {
    selector: ':selected',
    style: {
      'border-color': '#3b82f6',
      'border-width': 5,
      'z-index': 999
    }
  },

  {
    selector: 'node:active',
    style: {
      'overlay-opacity': 0.15,
      'overlay-color': '#3b82f6',
      'overlay-padding': 10
    }
  },

  {
    selector: '.dimmed',
    style: {
      'opacity': 0.25
    }
  }
];

// Layout configurations
export const enhancedLayoutConfigs = {
  dagre: {
    name: 'dagre',
    rankDir: 'TB',
    nodeSep: 120,
    rankSep: 180,
    animate: true,
    animationDuration: 600,
    animationEasing: 'ease-out-cubic'
  },

  concentric: {
    name: 'concentric',
    concentric: (node: any) => {
      const type = node.data('type');
      if (type === 'terraform_change') return 10;
      if (type === 'falco_event') return 1;
      return 5;
    },
    levelWidth: () => 2,
    minNodeSpacing: 120,
    animate: true,
    animationDuration: 800
  },

  cose: {
    name: 'cose',
    nodeRepulsion: 12000,
    idealEdgeLength: 150,
    animate: true,
    animationDuration: 1000,
    nodeOverlap: 30,
    gravity: 0.2
  },

  grid: {
    name: 'grid',
    rows: undefined,
    cols: undefined,
    animate: true,
    animationDuration: 500
  }
};

// Enhanced Cytoscape config
export const enhancedCytoscapeConfig = {
  style: enhancedCytoscapeStylesheet,
  minZoom: 0.2,
  maxZoom: 4,
  boxSelectionEnabled: true,
  autounselectify: false,
  autoungrabify: false,
  motionBlur: true,
  motionBlurOpacity: 0.2,
  pixelRatio: 'auto'
};
