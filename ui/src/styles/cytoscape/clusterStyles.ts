/**
 * Cytoscape Cluster and State Styles
 * Interaction states, highlights, and cluster visualization
 */

interface CytoscapeStyle {
  selector: string;
  style: Record<string, unknown>;
}

export const clusterStyles: CytoscapeStyle[] = [
  // Highlighted elements (path/selection)
  {
    selector: '.highlighted',
    style: {
      'background-color': '#fef3c7',
      'border-color': '#f59e0b',
      'border-width': 5,
      'line-color': '#f59e0b',
      'target-arrow-color': '#f59e0b',
      'width': 6,
      'z-index': 999,
    },
  },

  // Selected elements
  {
    selector: ':selected',
    style: {
      'border-color': '#3b82f6',
      'border-width': 5,
      'z-index': 999,
    },
  },

  // Active nodes
  {
    selector: 'node:active',
    style: {
      'overlay-opacity': 0.15,
      'overlay-color': '#3b82f6',
      'overlay-padding': 10,
    },
  },

  // Dimmed elements
  {
    selector: '.dimmed',
    style: {
      'opacity': 0.25,
    },
  },

  // Parent nodes styling (for compound nodes like VPC/Subnet)
  {
    selector: '$node > node',
    style: {
      'padding-top': '10px',
      'padding-left': '10px',
      'padding-bottom': '10px',
      'padding-right': '10px',
      'text-valign': 'top',
      'text-halign': 'left',
      'background-opacity': 0.2,
    },
  },

  // Severity-based styling
  {
    selector: 'node[severity="critical"]',
    style: {
      'border-width': 5,
      'border-color': '#c92a2a',
    },
  },

  {
    selector: 'node[severity="high"]',
    style: {
      'border-width': 4,
      'border-color': '#e8590c',
    },
  },

  {
    selector: 'node[severity="medium"]',
    style: {
      'border-width': 3,
      'border-color': '#fab005',
    },
  },

  {
    selector: 'node[severity="low"]',
    style: {
      'border-width': 2,
      'border-color': '#51cf66',
    },
  },
];
