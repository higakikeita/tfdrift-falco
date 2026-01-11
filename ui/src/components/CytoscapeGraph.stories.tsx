import type { Meta, StoryObj } from '@storybook/react';
import { CytoscapeGraph } from './CytoscapeGraph';
import {
  mockSmallGraph,
  mockMediumGraph,
  mockGraphWithDrift,
  mockAllResourceTypes,
  mockEmptyGraph,
  generateLargeGraph
} from '../mocks/graphData';

/**
 * CytoscapeGraph - TFDrift-Falco Graph Visualization Component
 *
 * AWS構成図を可視化するメインコンポーネント。
 * VPC/Subnet階層、Drift状態、複数のレイアウトアルゴリズムをサポート。
 */
const meta = {
  title: 'Components/CytoscapeGraph',
  component: CytoscapeGraph,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: `
TFDrift-FalcoのGraph Viewで使用されるCytoscape.jsベースの可視化コンポーネント。

## 主な機能
- **VPC/Subnet階層の可視化**: コンパウンドノードによる親子関係表示
- **複数レイアウト**: fcose（推奨）、dagre、cose、grid
- **スケール調整**: 0.5x〜2.0x
- **AWS公式アイコン**: 15種類以上のサービスアイコン
- **Drift状態の強調表示**: Critical/High/Medium severity

## 推奨レイアウト
- **fcose**: VPC/Subnet階層がある場合（デフォルト）
- **dagre**: 階層的な依存関係を見たい場合
- **cose**: 物理シミュレーションで配置
- **grid**: グリッド状に整列
        `
      }
    }
  },
  tags: ['autodocs'],
  argTypes: {
    layout: {
      control: 'select',
      options: ['fcose', 'dagre', 'cose', 'grid'],
      description: 'グラフレイアウトアルゴリズム',
      table: {
        defaultValue: { summary: 'fcose' }
      }
    },
    elements: {
      description: 'グラフデータ（ノードとエッジ）',
      control: false
    },
    onNodeClick: {
      description: 'ノードクリック時のコールバック',
      action: 'nodeClicked'
    },
    onEdgeClick: {
      description: 'エッジクリック時のコールバック',
      action: 'edgeClicked'
    },
    highlightedPath: {
      description: 'ハイライトするノードIDの配列',
      control: 'object'
    },
    className: {
      description: '追加のCSSクラス名',
      control: 'text'
    }
  },
  args: {
    onNodeClick: (nodeId: string, nodeData: any) => {
      console.log('Node clicked:', nodeId, nodeData);
    },
    onEdgeClick: (edgeId: string, edgeData: any) => {
      console.log('Edge clicked:', edgeId, edgeData);
    }
  }
} satisfies Meta<typeof CytoscapeGraph>;

export default meta;
type Story = StoryObj<typeof meta>;

// ==================== 基本 ====================

/**
 * デフォルト表示
 *
 * 10ノードの小規模グラフ。VPC/Subnet階層を含む。
 */
export const Default: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose'
  }
};

/**
 * 空データ
 *
 * ノードもエッジも存在しない状態。
 * エラーハンドリングの確認用。
 */
export const Empty: Story = {
  args: {
    elements: mockEmptyGraph,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: 'データが空の場合の表示。グラフが表示されない状態を確認できます。'
      }
    }
  }
};

// ==================== VPC/Subnet階層 ====================

/**
 * VPC/Subnet階層表示（推奨）
 *
 * fcoseレイアウトでVPC/Subnet階層を明確に表示。
 * これがTFDrift-Falcoの推奨構成です。
 */
export const WithVPCHierarchy: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: `
30ノードの中規模グラフ。VPC内に3つのSubnetがあり、各Subnet内にリソースが配置されています。

**階層構造:**
- VPC (1)
  - Subnet private-a (6 resources)
  - Subnet private-b (3 resources)
  - Subnet public-a (3 resources)
- VPCレベルリソース (8)
        `
      }
    }
  }
};

// ==================== レイアウトバリエーション ====================

/**
 * fcoseレイアウト（推奨）
 *
 * Force-directed compound spring embedder。
 * コンパウンドノード（VPC/Subnet）に最適化。
 */
export const LayoutFcose: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: 'VPC/Subnet階層がある場合の推奨レイアウト。物理シミュレーションベースで自然な配置。'
      }
    }
  }
};

/**
 * dagreレイアウト
 *
 * 階層的な有向グラフレイアウト。
 * 依存関係の方向性を明確にしたい場合に有効。
 */
export const LayoutDagre: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'dagre'
  },
  parameters: {
    docs: {
      description: {
        story: '上から下への階層的レイアウト。依存関係のフローが見やすい。'
      }
    }
  }
};

/**
 * coseレイアウト
 *
 * Compound Spring Embedder（物理シミュレーション）。
 */
export const LayoutCose: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'cose'
  },
  parameters: {
    docs: {
      description: {
        story: '物理シミュレーションベース。ノード間の力のバランスで配置。'
      }
    }
  }
};

/**
 * gridレイアウト
 *
 * グリッド状に整列。
 */
export const LayoutGrid: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'grid'
  },
  parameters: {
    docs: {
      description: {
        story: 'ノードをグリッド状に整列。規則的な配置が必要な場合に使用。'
      }
    }
  }
};

// ==================== データ量バリエーション ====================

/**
 * 小規模グラフ（10ノード）
 *
 * シンプルなグラフ。プロトタイピングやテストに最適。
 */
export const SmallGraph: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: '10ノードの小規模グラフ。VPC + 2 Subnets + リソース。'
      }
    }
  }
};

/**
 * 中規模グラフ（30ノード）
 *
 * 実際のプロジェクトに近いサイズ。
 */
export const MediumGraph: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: '30ノードの中規模グラフ。VPC + 3 Subnets + 複数リソース + エッジ。'
      }
    }
  }
};

/**
 * 大規模グラフ（100ノード）
 *
 * パフォーマンステスト用。
 */
export const LargeGraph: Story = {
  args: {
    elements: generateLargeGraph(100),
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: '100ノードの大規模グラフ。レンダリングパフォーマンスの確認に使用。'
      }
    }
  }
};

/**
 * 超大規模グラフ（200ノード）
 *
 * 極限のパフォーマンステスト。
 */
export const VeryLargeGraph: Story = {
  args: {
    elements: generateLargeGraph(200),
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: '200ノードの超大規模グラフ。fcoseレイアウトの限界テスト。'
      }
    }
  }
};

// ==================== Drift状態 ====================

/**
 * Drift強調表示
 *
 * Driftのあるリソースをボーダーカラーで強調。
 * - Critical (赤): Missing
 * - High (オレンジ): Modified
 * - Medium (黄): Unmanaged
 */
export const DriftHighlighted: Story = {
  args: {
    elements: mockGraphWithDrift,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: `
Drift状態を持つリソースの表示例。

**Severity:**
- \`critical\`: Missing resources（赤ボーダー）
- \`high\`: Modified resources（オレンジボーダー）
- \`medium\`: Unmanaged resources（黄ボーダー）
        `
      }
    }
  }
};

// ==================== 全AWSサービス ====================

/**
 * 全AWSサービスタイプ
 *
 * サポートしている全AWSサービスのアイコン表示。
 * アイコンとスタイルの確認用。
 */
export const AllResourceTypes: Story = {
  args: {
    elements: mockAllResourceTypes,
    layout: 'grid'
  },
  parameters: {
    docs: {
      description: {
        story: `
サポートしている全AWSサービスタイプ一覧：

**Compute:** EKS, ECS, Node Group, Addon
**Database:** RDS, ElastiCache
**Network:** VPC, Subnet, ALB, IGW, NAT, Route Table, Security Group
**Security:** IAM Role/Policy, KMS, Secrets Manager
**Monitoring:** CloudWatch Logs
        `
      }
    }
  }
};

// ==================== インタラクション ====================

/**
 * ノードクリック時のアクション
 *
 * ノードをクリックするとActionsパネルにログが表示されます。
 */
export const WithNodeClick: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose',
    onNodeClick: (nodeId: string, nodeData: any) => {
      console.log('Node clicked:', { nodeId, nodeData });
    }
  },
  parameters: {
    docs: {
      description: {
        story: 'ノードをクリックしてActionsパネルを確認してください。'
      }
    }
  }
};

/**
 * エッジクリック時のアクション
 *
 * エッジをクリックするとActionsパネルにログが表示されます。
 */
export const WithEdgeClick: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose',
    onEdgeClick: (edgeId: string, edgeData: any) => {
      console.log('Edge clicked:', { edgeId, edgeData });
    }
  },
  parameters: {
    docs: {
      description: {
        story: 'エッジ（矢印）をクリックしてActionsパネルを確認してください。'
      }
    }
  }
};

/**
 * パスハイライト
 *
 * 特定のノード間のパスをハイライト表示。
 */
export const WithHighlightedPath: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose',
    highlightedPath: ['eks-1', 'sg-1', 'iam-1']
  },
  parameters: {
    docs: {
      description: {
        story: 'highlightedPathで指定したノードとそれらを繋ぐエッジが強調表示されます。'
      }
    }
  }
};

// ==================== Playground ====================

/**
 * Playground
 *
 * すべてのControlsを使って自由に試せます。
 */
export const Playground: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose',
    onNodeClick: (nodeId: string, nodeData: any) => {
      console.log('Node clicked:', { nodeId, nodeData });
    },
    onEdgeClick: (edgeId: string, edgeData: any) => {
      console.log('Edge clicked:', { edgeId, edgeData });
    },
    highlightedPath: [],
    className: ''
  },
  parameters: {
    docs: {
      description: {
        story: 'Controlsパネルで各プロパティを変更して動作を確認できます。'
      }
    }
  }
};
