# グラフビジュアライゼーション アーキテクチャ設計 🎨

**目的**: 大規模インフラ（1000+ リソース）でも高速に動作する、Neo4j思想のグラフビジュアライゼーション

---

## 現在の課題 ⚠️

### React Flowの限界
- ✅ 小規模（~100ノード）では優れたUX
- ❌ 大規模（1000+ノード）でパフォーマンス劣化
- ❌ 階層構造とフラットな依存関係の両立が困難
- ❌ DOMベースのレンダリングでメモリ消費が大きい

### 必要な機能
1. **高速描画**: 1000+ノードで60fps維持
2. **柔軟なレイアウト**: 階層/力学/放射状など
3. **リアルタイム更新**: ドリフト検出時の即座の反映
4. **関係性探索**: 依存関係、影響範囲の可視化
5. **事前計算**: 複雑な関係性の事前処理

---

## 提案アーキテクチャ 🏗️

### 3層構造

```
┌─────────────────────────────────────────────────────────┐
│                  Layer 1: UI/UX (React)                  │
│  - インタラクション管理                                   │
│  - 状態管理 (Zustand)                                    │
│  - コンポーネント (検索、フィルター、詳細パネル)           │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│          Layer 2: 描画エンジン (Canvas/WebGL)            │
│  - sigma.js または vis-network                           │
│  - WebGLレンダラー (高速描画)                             │
│  - カスタムシェーダー (グロー、アニメーション)              │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│       Layer 3: グラフデータエンジン (Backend)             │
│  - Neo4j風のグラフモデル                                 │
│  - 関係性の事前計算                                      │
│  - パスファインディング                                   │
│  - 影響範囲分析                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 1. Neo4j系グラフDB思想 🔗

### グラフモデル

```go
// pkg/graph/model.go

// Node represents a graph node (Cypher風)
type Node struct {
    ID         string                 `json:"id"`
    Labels     []string               `json:"labels"`      // ["Resource", "EC2", "Drifted"]
    Properties map[string]interface{} `json:"properties"`
}

// Relationship represents a graph edge
type Relationship struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`        // "DEPENDS_ON", "CONTAINS", "CONNECTS_TO"
    StartNode  string                 `json:"start_node"`
    EndNode    string                 `json:"end_node"`
    Properties map[string]interface{} `json:"properties"`
}

// GraphDatabase in-memory graph database
type GraphDatabase struct {
    nodes         map[string]*Node
    relationships map[string]*Relationship

    // Indexes for fast lookups
    nodesByLabel  map[string][]*Node
    outgoing      map[string][]*Relationship  // node -> outgoing edges
    incoming      map[string][]*Relationship  // node -> incoming edges
}
```

### Cypher風のクエリインターフェース

```go
// Query examples (内部実装)

// Find all EC2 instances that depend on a VPC
// MATCH (vpc:VPC)-[:CONTAINS]->(subnet:Subnet)-[:CONTAINS]->(ec2:EC2)
// WHERE vpc.id = 'vpc-123' AND ec2.has_drift = true
// RETURN ec2

func (db *GraphDatabase) FindDriftedEC2InVPC(vpcID string) []*Node {
    vpc := db.GetNode(vpcID)
    if vpc == nil {
        return nil
    }

    results := []*Node{}

    // Traverse: VPC -> Subnet -> EC2
    for _, rel := range db.outgoing[vpcID] {
        if rel.Type == "CONTAINS" {
            subnet := db.nodes[rel.EndNode]
            if hasLabel(subnet, "Subnet") {
                for _, subnetRel := range db.outgoing[subnet.ID] {
                    if subnetRel.Type == "CONTAINS" {
                        ec2 := db.nodes[subnetRel.EndNode]
                        if hasLabel(ec2, "EC2") && ec2.Properties["has_drift"].(bool) {
                            results = append(results, ec2)
                        }
                    }
                }
            }
        }
    }

    return results
}

// Find impact radius of a change
// MATCH (changed:Resource)-[*1..3]-(affected:Resource)
// RETURN affected, length(path)

func (db *GraphDatabase) FindImpactRadius(nodeID string, maxDepth int) map[string]int {
    distances := make(map[string]int)
    queue := []string{nodeID}
    distances[nodeID] = 0

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        currentDist := distances[current]

        if currentDist >= maxDepth {
            continue
        }

        // BFS traversal
        for _, rel := range db.outgoing[current] {
            if _, visited := distances[rel.EndNode]; !visited {
                distances[rel.EndNode] = currentDist + 1
                queue = append(queue, rel.EndNode)
            }
        }
        for _, rel := range db.incoming[current] {
            if _, visited := distances[rel.StartNode]; !visited {
                distances[rel.StartNode] = currentDist + 1
                queue = append(queue, rel.StartNode)
            }
        }
    }

    return distances
}
```

### 関係性の種類（Relationship Types）

```go
const (
    // Hierarchical relationships
    CONTAINS         = "CONTAINS"          // VPC contains Subnet
    PART_OF          = "PART_OF"           // Subnet part of VPC

    // Dependency relationships
    DEPENDS_ON       = "DEPENDS_ON"        // EC2 depends on Subnet
    REQUIRED_BY      = "REQUIRED_BY"       // Subnet required by EC2

    // Network relationships
    CONNECTS_TO      = "CONNECTS_TO"       // EC2 connects to RDS
    ROUTES_TO        = "ROUTES_TO"         // Route table routes to Gateway

    // Security relationships
    ALLOWS           = "ALLOWS"            // SecurityGroup allows traffic
    BLOCKS           = "BLOCKS"            // NACL blocks traffic

    // Change relationships
    DRIFTED_FROM     = "DRIFTED_FROM"      // Current state drifted from desired
    CAUSED_DRIFT_IN  = "CAUSED_DRIFT_IN"   // Change A caused drift in B
)
```

---

## 2. Canvas/WebGL描画エンジン 🎨

### 選定: sigma.js (WebGL)

**理由**:
- ✅ WebGLベースで10万ノード以上をサポート
- ✅ カスタマイズ可能なレンダラー
- ✅ Reactとの統合が容易
- ✅ 力学レイアウト (force-directed) 標準搭載
- ✅ アクティブな開発コミュニティ

### 実装例

```typescript
// ui/src/components/graph/SigmaGraph.tsx

import { useEffect, useRef } from 'react';
import Sigma from 'sigma';
import Graph from 'graphology';
import { circular } from 'graphology-layout';
import forceAtlas2 from 'graphology-layout-forceatlas2';

interface SigmaGraphProps {
  graphData: {
    nodes: Array<{ id: string; label: string; type: string; x?: number; y?: number }>;
    edges: Array<{ id: string; source: string; target: string; type: string }>;
  };
}

export const SigmaGraph: React.FC<SigmaGraphProps> = ({ graphData }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const sigmaRef = useRef<Sigma | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    // Create graphology instance
    const graph = new Graph();

    // Add nodes
    graphData.nodes.forEach(node => {
      graph.addNode(node.id, {
        label: node.label,
        size: getNodeSize(node.type),
        color: getNodeColor(node.type),
        x: node.x ?? Math.random(),
        y: node.y ?? Math.random(),
      });
    });

    // Add edges
    graphData.edges.forEach(edge => {
      graph.addEdge(edge.source, edge.target, {
        size: 2,
        color: getEdgeColor(edge.type),
        type: getEdgeType(edge.type),
      });
    });

    // Apply layout
    if (!graphData.nodes[0]?.x) {
      // Use force-directed layout if positions not provided
      forceAtlas2.assign(graph, {
        iterations: 100,
        settings: {
          gravity: 1,
          scalingRatio: 10,
        }
      });
    }

    // Create Sigma instance
    const sigma = new Sigma(graph, containerRef.current, {
      renderEdgeLabels: false,
      enableEdgeEvents: true,
      // Custom node renderer for AWS resources
      defaultNodeType: 'circle',
      // WebGL renderer for performance
      labelRenderer: customLabelRenderer,
    });

    sigmaRef.current = sigma;

    // Add interactions
    sigma.on('clickNode', ({ node }) => {
      console.log('Clicked node:', node);
      // Show detail panel
    });

    sigma.on('enterNode', ({ node }) => {
      // Highlight connected nodes
      const neighbors = new Set(graph.neighbors(node));
      sigma.setSetting('nodeReducer', (n, data) => {
        if (n === node || neighbors.has(n)) {
          return { ...data, highlighted: true };
        }
        return { ...data, color: '#E2E2E2', highlighted: false };
      });
      sigma.refresh();
    });

    sigma.on('leaveNode', () => {
      sigma.setSetting('nodeReducer', null);
      sigma.refresh();
    });

    return () => {
      sigma.kill();
    };
  }, [graphData]);

  return (
    <div
      ref={containerRef}
      style={{ width: '100%', height: '100%', background: '#f8fafc' }}
    />
  );
};

// Helper functions
function getNodeSize(type: string): number {
  switch (type) {
    case 'vpc': return 30;
    case 'subnet': return 20;
    case 'ec2': return 15;
    default: return 10;
  }
}

function getNodeColor(type: string): string {
  switch (type) {
    case 'vpc': return '#3B82F6';
    case 'subnet': return '#10B981';
    case 'ec2': return '#F59E0B';
    case 'rds': return '#EF4444';
    default: return '#64748b';
  }
}

function getEdgeColor(type: string): string {
  switch (type) {
    case 'DEPENDS_ON': return '#3B82F6';
    case 'CONTAINS': return '#10B981';
    case 'CONNECTS_TO': return '#F59E0B';
    default: return '#CBD5E1';
  }
}

function getEdgeType(type: string): string {
  switch (type) {
    case 'DEPENDS_ON': return 'arrow';
    case 'CONTAINS': return 'line';
    default: return 'line';
  }
}

function customLabelRenderer(context: CanvasRenderingContext2D, data: any) {
  // Custom AWS icon rendering
  const { x, y, size, label } = data;

  // Draw icon (could be loaded from sprite sheet)
  context.fillStyle = data.color;
  context.beginPath();
  context.arc(x, y, size, 0, Math.PI * 2);
  context.fill();

  // Draw label
  context.fillStyle = '#000';
  context.font = '12px Arial';
  context.fillText(label, x + size + 5, y + 5);
}
```

---

## 3. 事前計算エンジン ⚡

### バックエンドで事前計算する情報

```go
// pkg/graph/precompute.go

type PrecomputedData struct {
    // Dependency chains
    DependencyTree    map[string][]string       `json:"dependency_tree"`

    // Impact analysis
    ImpactRadius      map[string]map[string]int `json:"impact_radius"`  // node -> {affected_node: distance}

    // Critical paths
    CriticalPaths     [][]string                `json:"critical_paths"`

    // Clustering
    Communities       map[string]int            `json:"communities"`     // node -> community_id

    // Centrality metrics
    PageRank          map[string]float64        `json:"pagerank"`
    BetweennessCentrality map[string]float64    `json:"betweenness"`

    // Layout positions (pre-computed for common views)
    LayoutPositions   map[string]Position       `json:"layout_positions"`
}

type Position struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

// PrecomputeEngine performs expensive calculations
type PrecomputeEngine struct {
    db    *GraphDatabase
    cache *PrecomputedData
}

func (e *PrecomputeEngine) ComputeAll() *PrecomputedData {
    log.Info("Starting graph precomputation...")

    start := time.Now()

    data := &PrecomputedData{
        DependencyTree:        e.computeDependencyTree(),
        ImpactRadius:          e.computeImpactRadius(),
        CriticalPaths:         e.findCriticalPaths(),
        Communities:           e.detectCommunities(),
        PageRank:              e.computePageRank(),
        BetweennessCentrality: e.computeBetweenness(),
        LayoutPositions:       e.computeLayout(),
    }

    duration := time.Since(start)
    log.Infof("Graph precomputation completed in %v", duration)

    return data
}

// Dependency tree: Find all transitive dependencies
func (e *PrecomputeEngine) computeDependencyTree() map[string][]string {
    tree := make(map[string][]string)

    for nodeID := range e.db.nodes {
        deps := e.findAllDependencies(nodeID, 10) // max depth 10
        tree[nodeID] = deps
    }

    return tree
}

func (e *PrecomputeEngine) findAllDependencies(nodeID string, maxDepth int) []string {
    visited := make(map[string]bool)
    queue := []struct {
        id    string
        depth int
    }{{nodeID, 0}}

    var deps []string

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        if current.depth >= maxDepth {
            continue
        }

        if visited[current.id] {
            continue
        }
        visited[current.id] = true

        // Find dependencies
        for _, rel := range e.db.outgoing[current.id] {
            if rel.Type == "DEPENDS_ON" {
                deps = append(deps, rel.EndNode)
                queue = append(queue, struct {
                    id    string
                    depth int
                }{rel.EndNode, current.depth + 1})
            }
        }
    }

    return deps
}

// Impact radius: How far does a change propagate?
func (e *PrecomputeEngine) computeImpactRadius() map[string]map[string]int {
    impact := make(map[string]map[string]int)

    for nodeID := range e.db.nodes {
        impact[nodeID] = e.db.FindImpactRadius(nodeID, 5) // 5 hops
    }

    return impact
}

// Critical paths: Paths with highest dependency chains
func (e *PrecomputeEngine) findCriticalPaths() [][]string {
    // Find paths where many resources depend on a single point
    var paths [][]string

    for nodeID := range e.db.nodes {
        inDegree := len(e.db.incoming[nodeID])
        if inDegree > 3 { // Critical if >3 dependents
            path := e.tracePath(nodeID)
            paths = append(paths, path)
        }
    }

    return paths
}

func (e *PrecomputeEngine) tracePath(nodeID string) []string {
    path := []string{nodeID}
    current := nodeID

    // Follow DEPENDS_ON chain
    for len(e.db.outgoing[current]) > 0 {
        for _, rel := range e.db.outgoing[current] {
            if rel.Type == "DEPENDS_ON" {
                path = append(path, rel.EndNode)
                current = rel.EndNode
                break
            }
        }
    }

    return path
}

// Community detection using simple label propagation
func (e *PrecomputeEngine) detectCommunities() map[string]int {
    // Initialize each node with unique community
    communities := make(map[string]int)
    nodeList := make([]string, 0, len(e.db.nodes))

    for nodeID := range e.db.nodes {
        communities[nodeID] = len(nodeList)
        nodeList = append(nodeList, nodeID)
    }

    // Label propagation
    for iteration := 0; iteration < 10; iteration++ {
        changed := false

        for _, nodeID := range nodeList {
            // Count neighbor communities
            counts := make(map[int]int)

            for _, rel := range e.db.outgoing[nodeID] {
                counts[communities[rel.EndNode]]++
            }
            for _, rel := range e.db.incoming[nodeID] {
                counts[communities[rel.StartNode]]++
            }

            // Adopt most common community
            maxCount := 0
            maxCommunity := communities[nodeID]
            for community, count := range counts {
                if count > maxCount {
                    maxCount = count
                    maxCommunity = community
                }
            }

            if communities[nodeID] != maxCommunity {
                communities[nodeID] = maxCommunity
                changed = true
            }
        }

        if !changed {
            break
        }
    }

    return communities
}

// PageRank: Importance ranking
func (e *PrecomputeEngine) computePageRank() map[string]float64 {
    damping := 0.85
    iterations := 20

    ranks := make(map[string]float64)
    newRanks := make(map[string]float64)

    // Initialize
    n := len(e.db.nodes)
    initialRank := 1.0 / float64(n)
    for nodeID := range e.db.nodes {
        ranks[nodeID] = initialRank
    }

    // Iterate
    for i := 0; i < iterations; i++ {
        for nodeID := range e.db.nodes {
            sum := 0.0

            // Sum contributions from incoming nodes
            for _, rel := range e.db.incoming[nodeID] {
                outDegree := len(e.db.outgoing[rel.StartNode])
                if outDegree > 0 {
                    sum += ranks[rel.StartNode] / float64(outDegree)
                }
            }

            newRanks[nodeID] = (1-damping)/float64(n) + damping*sum
        }

        ranks, newRanks = newRanks, ranks
    }

    return ranks
}

// Betweenness centrality: How often a node appears on shortest paths
func (e *PrecomputeEngine) computeBetweenness() map[string]float64 {
    betweenness := make(map[string]float64)

    // For each pair of nodes, find shortest path
    for startID := range e.db.nodes {
        paths := e.findShortestPaths(startID)

        for _, path := range paths {
            // Increment betweenness for intermediate nodes
            for i := 1; i < len(path)-1; i++ {
                betweenness[path[i]]++
            }
        }
    }

    // Normalize
    n := float64(len(e.db.nodes))
    for nodeID := range betweenness {
        betweenness[nodeID] /= (n - 1) * (n - 2) / 2
    }

    return betweenness
}

func (e *PrecomputeEngine) findShortestPaths(startID string) [][]string {
    // BFS to find shortest paths
    distances := make(map[string]int)
    parents := make(map[string][]string)

    queue := []string{startID}
    distances[startID] = 0

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        currentDist := distances[current]

        for _, rel := range e.db.outgoing[current] {
            neighbor := rel.EndNode

            if _, visited := distances[neighbor]; !visited {
                distances[neighbor] = currentDist + 1
                parents[neighbor] = []string{current}
                queue = append(queue, neighbor)
            } else if distances[neighbor] == currentDist+1 {
                parents[neighbor] = append(parents[neighbor], current)
            }
        }
    }

    // Reconstruct paths
    var paths [][]string
    for endID := range e.db.nodes {
        if endID != startID {
            nodePaths := e.reconstructPaths(startID, endID, parents)
            paths = append(paths, nodePaths...)
        }
    }

    return paths
}

func (e *PrecomputeEngine) reconstructPaths(start, end string, parents map[string][]string) [][]string {
    if start == end {
        return [][]string{{start}}
    }

    parentList, exists := parents[end]
    if !exists {
        return nil
    }

    var paths [][]string
    for _, parent := range parentList {
        subPaths := e.reconstructPaths(start, parent, parents)
        for _, subPath := range subPaths {
            path := append(subPath, end)
            paths = append(paths, path)
        }
    }

    return paths
}

// Pre-compute layout positions
func (e *PrecomputeEngine) computeLayout() map[string]Position {
    positions := make(map[string]Position)

    // Use force-directed layout algorithm
    // This is a simplified version - in production use proper force simulation

    nodes := make([]string, 0, len(e.db.nodes))
    for nodeID := range e.db.nodes {
        nodes = append(nodes, nodeID)
        // Random initial position
        positions[nodeID] = Position{
            X: rand.Float64() * 1000,
            Y: rand.Float64() * 1000,
        }
    }

    // Simulate forces
    for iteration := 0; iteration < 100; iteration++ {
        forces := make(map[string]Position)

        // Repulsive force between all nodes
        for i, node1 := range nodes {
            for j, node2 := range nodes {
                if i >= j {
                    continue
                }

                dx := positions[node2].X - positions[node1].X
                dy := positions[node2].Y - positions[node1].Y
                dist := math.Sqrt(dx*dx + dy*dy)

                if dist < 1 {
                    dist = 1
                }

                // Coulomb's law
                force := 1000 / (dist * dist)
                fx := force * dx / dist
                fy := force * dy / dist

                f1 := forces[node1]
                f1.X -= fx
                f1.Y -= fy
                forces[node1] = f1

                f2 := forces[node2]
                f2.X += fx
                f2.Y += fy
                forces[node2] = f2
            }
        }

        // Attractive force along edges
        for _, rel := range e.db.relationships {
            start := positions[rel.StartNode]
            end := positions[rel.EndNode]

            dx := end.X - start.X
            dy := end.Y - start.Y
            dist := math.Sqrt(dx*dx + dy*dy)

            if dist < 1 {
                dist = 1
            }

            // Hooke's law
            force := dist * 0.01
            fx := force * dx / dist
            fy := force * dy / dist

            f1 := forces[rel.StartNode]
            f1.X += fx
            f1.Y += fy
            forces[rel.StartNode] = f1

            f2 := forces[rel.EndNode]
            f2.X -= fx
            f2.Y -= fy
            forces[rel.EndNode] = f2
        }

        // Apply forces
        for nodeID, force := range forces {
            pos := positions[nodeID]
            pos.X += force.X * 0.1
            pos.Y += force.Y * 0.1
            positions[nodeID] = pos
        }
    }

    return positions
}
```

### API エンドポイント

```go
// pkg/api/handlers/graph_precomputed.go

// GetPrecomputedData returns pre-calculated graph data
func (h *GraphHandler) GetPrecomputedData(c *gin.Context) {
    data := h.precomputeEngine.GetCache()

    if data == nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "success": false,
            "error":   "Precomputed data not available yet",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    data,
    })
}

// GetImpactAnalysis returns impact radius for a specific node
func (h *GraphHandler) GetImpactAnalysis(c *gin.Context) {
    nodeID := c.Param("id")

    data := h.precomputeEngine.GetCache()
    if data == nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "success": false,
            "error":   "Precomputed data not available",
        })
        return
    }

    impact, exists := data.ImpactRadius[nodeID]
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "Node not found",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "node_id":        nodeID,
            "affected_nodes": impact,
            "total_affected": len(impact),
        },
    })
}

// GetDependencyChain returns dependency chain for a node
func (h *GraphHandler) GetDependencyChain(c *gin.Context) {
    nodeID := c.Param("id")

    data := h.precomputeEngine.GetCache()
    if data == nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "success": false,
            "error":   "Precomputed data not available",
        })
        return
    }

    deps, exists := data.DependencyTree[nodeID]
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "Node not found",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "node_id":      nodeID,
            "dependencies": deps,
            "depth":        len(deps),
        },
    })
}
```

---

## 4. フロントエンド統合 🎯

### Zustand状態管理

```typescript
// ui/src/store/graphStore.ts

import create from 'zustand';
import { GraphData, PrecomputedData } from '../types/graph';

interface GraphStore {
  // Graph data
  graphData: GraphData | null;
  precomputedData: PrecomputedData | null;

  // UI state
  selectedNode: string | null;
  highlightedNodes: Set<string>;
  filter: {
    types: Set<string>;
    severity: Set<string>;
    hasFilter: boolean;
  };

  // Actions
  setGraphData: (data: GraphData) => void;
  setPrecomputedData: (data: PrecomputedData) => void;
  selectNode: (nodeId: string | null) => void;
  highlightImpactRadius: (nodeId: string) => void;
  clearHighlight: () => void;
  setFilter: (filter: Partial<GraphStore['filter']>) => void;
}

export const useGraphStore = create<GraphStore>((set, get) => ({
  graphData: null,
  precomputedData: null,
  selectedNode: null,
  highlightedNodes: new Set(),
  filter: {
    types: new Set(),
    severity: new Set(),
    hasFilter: false,
  },

  setGraphData: (data) => set({ graphData: data }),

  setPrecomputedData: (data) => set({ precomputedData: data }),

  selectNode: (nodeId) => set({ selectedNode: nodeId }),

  highlightImpactRadius: (nodeId) => {
    const { precomputedData } = get();
    if (!precomputedData) return;

    const impact = precomputedData.impactRadius[nodeId] || {};
    const affected = new Set(Object.keys(impact));
    affected.add(nodeId);

    set({ highlightedNodes: affected, selectedNode: nodeId });
  },

  clearHighlight: () => set({ highlightedNodes: new Set() }),

  setFilter: (newFilter) => set((state) => ({
    filter: {
      ...state.filter,
      ...newFilter,
      hasFilter: true,
    }
  })),
}));
```

---

## 5. 実装ロードマップ 📅

### Phase 1: バックエンドグラフエンジン (Week 1-2)
- [ ] GraphDatabase実装
- [ ] Neo4j風のグラフモデル
- [ ] 基本的なトラバーサルAPI
- [ ] 関係性タイプの定義

### Phase 2: 事前計算エンジン (Week 2-3)
- [ ] PrecomputeEngine実装
- [ ] 依存関係ツリー計算
- [ ] 影響範囲分析
- [ ] PageRank計算
- [ ] レイアウト事前計算

### Phase 3: sigma.js統合 (Week 3-4)
- [ ] sigma.jsセットアップ
- [ ] カスタムレンダラー
- [ ] AWS アイコン描画
- [ ] インタラクション実装

### Phase 4: Zustand状態管理 (Week 4)
- [ ] グラフストア実装
- [ ] フィルタリング機能
- [ ] ハイライト機能
- [ ] 検索機能

### Phase 5: 最適化 (Week 5)
- [ ] WebWorkerで事前計算
- [ ] 段階的レンダリング
- [ ] ビューポートベースの描画
- [ ] LOD (Level of Detail)

---

## 6. パフォーマンス目標 🎯

| 指標 | 目標 | 現状 (React Flow) |
|------|------|-------------------|
| 初回描画 | < 1秒 (1000ノード) | ~5秒 |
| FPS | 60fps (パン/ズーム) | 30-40fps |
| メモリ使用量 | < 500MB (10000ノード) | ~2GB |
| インタラクション応答 | < 100ms | 200-300ms |
| 事前計算 | < 5秒 (バックグラウンド) | N/A |

---

## 7. 段階的移行戦略 🔄

### Step 1: 並行実装
```typescript
// 両方のエンジンをサポート
<GraphView engine={useWebGL ? 'sigma' : 'react-flow'} />
```

### Step 2: 機能比較
- React Flow: 小規模（<100ノード）、開発環境
- sigma.js: 大規模（>100ノード）、本番環境

### Step 3: 完全移行
- sigma.jsをデフォルトに
- React Flowはレガシーサポートのみ

---

**次のステップ**: Phase 1の実装を開始

