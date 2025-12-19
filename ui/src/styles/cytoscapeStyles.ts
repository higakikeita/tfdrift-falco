/**
 * Cytoscape.js Style Definitions
 *
 * TFDrift-Falco因果関係グラフのビジュアルスタイル
 */

export const cytoscapeStylesheet: any[] = [
  // ==================== ノードスタイル ====================

  // Terraform Change nodes - 起点（紫のアイコン）
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
      'font-size': '9px',
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
      'font-size': '9px',
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
      'font-size': '8px',
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
      'font-size': '8px',
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
      'font-size': '8px',
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
      'font-size': '9px',
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
  }
];

// レイアウト設定
export const layoutConfigs = {
  dagre: {
    name: 'dagre',
    rankDir: 'TB',  // Top to Bottom
    nodeSep: 100,
    rankSep: 150,
    animate: true,
    animationDuration: 500,
    animationEasing: 'ease-out'
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
  autoungrabify: false
};
