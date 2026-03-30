/**
 * Cytoscape Edge Styles
 * Relationship and connection styling
 */

interface CytoscapeStyle {
  selector: string;
  style: Record<string, unknown>;
}

export const edgeStyles: CytoscapeStyle[] = [
  // Default edge style
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
      'text-background-padding': '3px',
    },
  },

  // Relationship types with specific colors
  {
    selector: 'edge[type="caused_by"]',
    style: {
      'line-color': '#ff6b6b',
      'target-arrow-color': '#ff6b6b',
      'width': 3,
    },
  },

  {
    selector: 'edge[type="grants_access"]',
    style: {
      'line-color': '#4dabf7',
      'target-arrow-color': '#4dabf7',
    },
  },

  {
    selector: 'edge[type="used_by"]',
    style: {
      'line-color': '#51cf66',
      'target-arrow-color': '#51cf66',
    },
  },

  {
    selector: 'edge[type="contains"]',
    style: {
      'line-color': '#ff922b',
      'target-arrow-color': '#ff922b',
      'width': 2,
    },
  },

  {
    selector: 'edge[type="triggered"]',
    style: {
      'line-color': '#f06595',
      'target-arrow-color': '#f06595',
      'width': 4,
    },
  },

  {
    selector: 'edge[type="depends_on"]',
    style: {
      'line-color': '#5c7cfa',
      'target-arrow-color': '#5c7cfa',
      'line-style': 'dashed',
    },
  },

  // Hover highlight
  {
    selector: 'edge.hover-highlight',
    style: {
      'width': 4,
      'opacity': 1,
    },
  },
];
