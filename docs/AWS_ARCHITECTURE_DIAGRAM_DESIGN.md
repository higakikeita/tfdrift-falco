# AWS標準構成図に準拠したdriftwire UI設計 🏗️

**参考**: [AWS アーキテクチャ図の描き方](https://aws.amazon.com/jp/builders-flash/202204/way-to-draw-architecture/)

---

## 設計思想

> **「何を伝えたいか」が最重要**
>
> driftwireでは「ドリフトがどこで発生し、なぜ起きたか」を伝えることが目的

---

## 1. 階層構造の実装 🏛️

### AWS標準の階層

```
AWS Global Infrastructure
  └── Region (us-east-1)
       └── VPC (10.0.0.0/16)
            ├── Availability Zone A
            │    ├── Public Subnet (10.0.1.0/24)
            │    │    ├── EC2 Instance
            │    │    └── NAT Gateway
            │    └── Private Subnet (10.0.2.0/24)
            │         ├── EC2 Instance
            │         └── RDS Instance
            └── Availability Zone B
                 ├── Public Subnet (10.0.3.0/24)
                 └── Private Subnet (10.0.4.0/24)
```

### React Flow実装

```typescript
// 階層的なグループノード
interface HierarchicalGroup {
  // Level 1: Region
  region: {
    id: 'region-us-east-1',
    label: 'us-east-1',
    type: 'region-group',
    style: {
      background: '#FFF7ED', // オレンジ系の薄い背景
      border: '3px solid #F59E0B',
      borderRadius: '16px',
      padding: '60px 40px',
    }
  },

  // Level 2: VPC
  vpc: {
    id: 'vpc-prod',
    label: 'Production VPC (10.0.0.0/16)',
    parentNode: 'region-us-east-1',
    type: 'vpc-group',
    style: {
      background: '#EFF6FF', // 青系の薄い背景
      border: '2px solid #3B82F6',
      borderRadius: '12px',
      padding: '40px 30px',
    }
  },

  // Level 3: Availability Zone
  az: {
    id: 'az-us-east-1a',
    label: 'Availability Zone A',
    parentNode: 'vpc-prod',
    type: 'az-group',
    style: {
      background: '#F0FDF4', // 緑系の薄い背景
      border: '2px dashed #10B981',
      borderRadius: '8px',
      padding: '30px 20px',
    }
  },

  // Level 4: Subnet
  subnet: {
    id: 'subnet-public-1a',
    label: 'Public Subnet (10.0.1.0/24)',
    parentNode: 'az-us-east-1a',
    type: 'subnet-group',
    style: {
      background: '#F0F9FF',
      border: '1px solid #0EA5E9',
      borderRadius: '6px',
      padding: '20px',
    }
  }
}
```

---

## 2. ネットワーク境界の明確化 🌐

### 境界の種類

```typescript
const networkBoundaries = {
  // VPC外（リージョンレベル）
  vpcExternal: {
    resources: ['S3', 'DynamoDB', 'CloudFront', 'Route53'],
    style: {
      background: '#FEF3C7', // 黄色系
      label: 'VPC外リソース',
    }
  },

  // インターネットゲートウェイ
  internetGateway: {
    position: 'vpc-boundary',
    icon: 'IGW',
    style: {
      border: '3px solid #059669',
    }
  },

  // パブリックサブネット
  publicSubnet: {
    internetAccess: true,
    style: {
      background: '#DBEAFE', // 明るい青
      icon: '🌐',
    }
  },

  // プライベートサブネット
  privateSubnet: {
    internetAccess: false,
    style: {
      background: '#E0E7FF', // 紫がかった青
      icon: '🔒',
    }
  }
}
```

### ネットワークフロー表現

```typescript
// インターネット → IGW → Public → NAT → Private
const networkFlow = [
  {
    from: 'internet',
    to: 'igw',
    label: '0.0.0.0/0',
    style: {
      stroke: '#10B981',
      strokeWidth: 3,
      animated: true,
    }
  },
  {
    from: 'igw',
    to: 'public-subnet',
    label: 'インバウンド',
    style: {
      stroke: '#3B82F6',
      strokeWidth: 2,
    }
  },
  {
    from: 'nat',
    to: 'private-subnet',
    label: 'アウトバウンド',
    style: {
      stroke: '#8B5CF6',
      strokeWidth: 2,
      strokeDasharray: '5,5',
    }
  }
];
```

---

## 3. 詳細度の切り替え機能 🔍

### 3つの表示モード

```typescript
type DetailLevel = 'overview' | 'standard' | 'detailed';

const displayConfig: Record<DetailLevel, DisplayConfig> = {
  // 概要モード: リージョン/VPC構造のみ
  overview: {
    showResources: false,
    showSubnets: false,
    showSecurityGroups: false,
    showMetadata: false,
    groupBy: 'vpc',
    nodeSize: 'small',
  },

  // 標準モード: 主要リソースを表示
  standard: {
    showResources: true,
    showSubnets: true,
    showSecurityGroups: false,
    showMetadata: 'minimal', // リソース名のみ
    groupBy: 'subnet',
    nodeSize: 'medium',
  },

  // 詳細モード: すべての情報を表示
  detailed: {
    showResources: true,
    showSubnets: true,
    showSecurityGroups: true,
    showMetadata: 'full', // スペック、タグ、設定
    groupBy: 'security-group',
    nodeSize: 'large',
  }
}
```

### ズームに応じた自動切り替え

```typescript
function onZoomChange(zoom: number) {
  if (zoom < 0.5) {
    setDetailLevel('overview');
  } else if (zoom < 1.0) {
    setDetailLevel('standard');
  } else {
    setDetailLevel('detailed');
  }
}
```

---

## 4. AWS公式アイコンの使用 🎨

### アイコンライブラリ

```typescript
// AWS Architecture Iconsを使用
import {
  SiAmazonec2,
  SiAmazons3,
  SiAmazonrds,
  SiAwslambda,
  // ... その他のAWSサービス
} from 'react-icons/si';

// カスタムアイコンマッピング
const awsIcons = {
  'aws_instance': SiAmazonec2,
  'aws_s3_bucket': SiAmazons3,
  'aws_db_instance': SiAmazonrds,
  'aws_lambda_function': SiAwslambda,
  'aws_vpc': VPCIcon,                    // カスタムSVG
  'aws_subnet': SubnetIcon,              // カスタムSVG
  'aws_internet_gateway': IGWIcon,       // カスタムSVG
  'aws_nat_gateway': NATIcon,            // カスタムSVG
  'aws_security_group': SecurityGroupIcon, // カスタムSVG
}
```

### アイコンサイズとスタイル

```typescript
const iconConfig = {
  overview: {
    size: 32,      // 小さいアイコン
    showLabel: false,
  },
  standard: {
    size: 48,      // 中サイズ
    showLabel: true,
    labelPosition: 'bottom',
  },
  detailed: {
    size: 64,      // 大きいアイコン
    showLabel: true,
    showBadge: true, // ドリフトバッジ
    showMetadata: true, // インスタンスタイプなど
  }
}
```

---

## 5. マルチAZ構成の表現 🏢

### 横並びレイアウト

```
┌─────────────────────────────────────────────────────────┐
│ VPC: Production (10.0.0.0/16)                            │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌────────── AZ-A ─────────┐  ┌────────── AZ-B ────────┐│
│  │                          │  │                         ││
│  │  [Public Subnet]        │  │  [Public Subnet]       ││
│  │    • Bastion            │  │    • Bastion           ││
│  │    • NAT Gateway        │  │    • NAT Gateway       ││
│  │                          │  │                         ││
│  │  [Private Subnet]       │  │  [Private Subnet]      ││
│  │    • Web Server         │  │    • Web Server        ││
│  │    • App Server         │  │    • App Server        ││
│  │                          │  │                         ││
│  │  [Data Subnet]          │  │  [Data Subnet]         ││
│  │    • RDS Primary        │  │    • RDS Standby       ││
│  │                          │  │                         ││
│  └──────────────────────────┘  └────────────────────────┘│
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### React Flow実装

```typescript
// AZを横に並べる自動レイアウト
function layoutMultiAZ(azNodes: Node[]) {
  const azWidth = 400;
  const spacing = 50;

  azNodes.forEach((az, index) => {
    az.position = {
      x: index * (azWidth + spacing),
      y: 100, // VPC内での固定Y位置
    };
  });

  return azNodes;
}
```

---

## 6. セキュリティグループの視覚化 🛡️

### セキュリティグループの表現方法

```typescript
// Option A: ノードの枠線色で区別
const securityGroupStyles = {
  'sg-web': {
    border: '3px solid #EF4444', // 赤: インターネット公開
    badge: '🌐 Public',
  },
  'sg-app': {
    border: '3px solid #F59E0B', // オレンジ: 内部通信
    badge: '🔒 Internal',
  },
  'sg-db': {
    border: '3px solid #10B981', // 緑: データベース
    badge: '🗄️ Database',
  }
}

// Option B: セキュリティグループを点線で囲む
const securityGroupBoundary = {
  id: 'sg-web-boundary',
  type: 'security-group-boundary',
  style: {
    border: '2px dashed #EF4444',
    borderRadius: '12px',
    background: 'rgba(239, 68, 68, 0.05)',
  },
  containedNodes: ['ec2-web-1', 'ec2-web-2']
}
```

### インバウンド/アウトバウンドルール

```typescript
// エッジでセキュリティルールを表現
const securityRules = [
  {
    from: 'internet',
    to: 'sg-web',
    label: 'Port 443 (HTTPS)',
    style: {
      stroke: '#EF4444',
      strokeWidth: 2,
      label: '0.0.0.0/0 → :443',
    }
  },
  {
    from: 'sg-web',
    to: 'sg-app',
    label: 'Port 8080 (App)',
    style: {
      stroke: '#F59E0B',
      strokeWidth: 2,
      label: 'sg-web → :8080',
    }
  }
];
```

---

## 7. ドリフト表現の統合 ⚠️

### AWS標準構成図 + ドリフト情報

```typescript
// ドリフトしたリソースの視覚的強調
const driftVisualization = {
  // 通常のリソース
  normal: {
    opacity: 0.7,           // やや薄く表示
    border: '2px solid #D1D5DB', // グレー
  },

  // ドリフト検出リソース
  drifted: {
    opacity: 1.0,           // 完全に表示
    border: '4px solid #EF4444', // 太い赤枠
    animation: 'pulse',     // パルスアニメーション
    badge: {
      icon: '⚠️',
      label: 'ドリフト',
      color: '#EF4444',
    },
    glow: '0 0 20px rgba(239, 68, 68, 0.5)', // 赤いグロー
  }
}
```

### タイムライン統合

```typescript
// 構成図 + タイムライン
interface ArchitectureDiagramWithTimeline {
  diagram: {
    nodes: Node[];
    edges: Edge[];
    layout: 'aws-standard';
  },
  timeline: {
    events: DriftEvent[];
    currentTime: Date;
    playback: boolean; // 時系列再生
  }
}

// 特定の時点の構成を表示
function showArchitectureAtTime(timestamp: Date) {
  // その時点のTerraform Stateを復元
  const stateAtTime = getStateAtTime(timestamp);
  // ドリフトイベントを重ねる
  const driftsUntilTime = getDriftsUntilTime(timestamp);
  // グラフを更新
  updateDiagram(stateAtTime, driftsUntilTime);
}
```

---

## 8. 実装計画の改訂 📋

### Phase 1: AWS標準階層構造 (Week 1)
- [ ] Region/VPC/AZ/Subnetの階層グループ
- [ ] React Flow Parent Node実装
- [ ] 自動レイアウトアルゴリズム
- [ ] マルチAZ横並びレイアウト

### Phase 2: ネットワーク境界 (Week 1-2)
- [ ] VPC内/外の視覚的区別
- [ ] インターネットゲートウェイ表現
- [ ] ネットワークフロー矢印
- [ ] パブリック/プライベート色分け

### Phase 3: AWS公式アイコン (Week 2)
- [ ] react-iconsからAWSアイコン統合
- [ ] カスタムSVGアイコン作成
- [ ] 詳細度別のアイコンサイズ
- [ ] バッジとラベル

### Phase 4: セキュリティグループ (Week 2-3)
- [ ] セキュリティグループ境界
- [ ] インバウンド/アウトバウンドルール
- [ ] ポート番号とプロトコル表示
- [ ] セキュリティリスクハイライト

### Phase 5: 詳細度切り替え (Week 3)
- [ ] 概要/標準/詳細の3モード
- [ ] ズームに応じた自動切り替え
- [ ] 情報の段階的表示
- [ ] パフォーマンス最適化

### Phase 6: ドリフト統合 (Week 3-4)
- [ ] ドリフトリソースの強調表示
- [ ] タイムライン連携
- [ ] 時系列再生機能
- [ ] 変更前後の比較

---

## 9. 参考実装例 💡

### AWS Well-Architected構成図スタイル

```typescript
const wellArchitectedStyle = {
  colors: {
    region: '#FF9900',    // AWSオレンジ
    vpc: '#3B82F6',       // 青
    publicSubnet: '#10B981',  // 緑
    privateSubnet: '#8B5CF6', // 紫
    securityGroup: '#EF4444', // 赤
  },

  typography: {
    fontFamily: 'Amazon Ember, Arial, sans-serif',
    labels: {
      region: { size: 24, weight: 'bold' },
      vpc: { size: 18, weight: 'bold' },
      subnet: { size: 14, weight: 'medium' },
      resource: { size: 12, weight: 'normal' },
    }
  },

  spacing: {
    regionPadding: 60,
    vpcPadding: 40,
    azPadding: 30,
    subnetPadding: 20,
    resourceMargin: 10,
  }
}
```

---

## 10. 成功指標 🎯

### ユーザビリティ目標
- ✅ 3秒以内にVPC構造を理解できる
- ✅ 5秒以内にドリフトリソースを特定できる
- ✅ 10秒以内にネットワークフローを把握できる

### 技術目標
- ✅ 100リソースで60fps維持
- ✅ ズーム/パンが滑らか
- ✅ 階層構造の自動レイアウト

---

**次のステップ**: Phase 1の実装を開始
