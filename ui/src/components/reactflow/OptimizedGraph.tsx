/**
 * Optimized Graph Component
 * Integrates LOD rendering, clustering, progressive loading for 1000+ nodes
 */

import { memo, useMemo, useCallback, useState } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  type Node,
  type Edge,
  type NodeTypes,
  ConnectionLineType,
  MarkerType,
  useNodesState,
  useEdgesState,
  type OnNodesChange,
  type OnEdgesChange,
} from 'reactflow';
import 'reactflow/dist/style.css';

import { LODNode, shouldUseLOD } from './LODNode';
import { ClusterNode } from './ClusterNode';
import { CustomNode } from './CustomNode';
import { useProgressiveGraph } from '../../hooks/useProgressiveGraph';
import { useGraphClustering } from '../../utils/graphClustering';
import { useDebounce } from '../../utils/memoryOptimization';

interface OptimizedGraphProps {
  nodes: Node[];
  edges: Edge[];
  onNodeClick?: (node: Node) => void;
  onEdgeClick?: (edge: Edge) => void;
  enableClustering?: boolean;
  enableProgressiveLoading?: boolean;
  enableLOD?: boolean;
  clusteringOptions?: {
    groupBy: 'type' | 'provider' | 'severity';
    minClusterSize?: number;
    maxClusterSize?: number;
  };
  progressiveOptions?: {
    batchSize?: number;
    batchDelay?: number;
  };
}

// Define custom node types
const nodeTypes: NodeTypes = {
  custom: CustomNode,
  lod: LODNode,
  cluster: ClusterNode,
};

// Default edge style
const defaultEdgeOptions = {
  type: ConnectionLineType.SmoothStep,
  animated: false,
  markerEnd: {
    type: MarkerType.ArrowClosed,
    width: 20,
    height: 20,
  },
  style: {
    strokeWidth: 2,
    stroke: '#94a3b8', // slate-400
  },
};

export const OptimizedGraph = memo(({
  nodes: initialNodes,
  edges: initialEdges,
  onNodeClick,
  onEdgeClick,
  enableClustering = true,
  enableProgressiveLoading = true,
  enableLOD = true,
  clusteringOptions = {
    groupBy: 'provider',
    minClusterSize: 5,
    maxClusterSize: 50,
  },
  progressiveOptions = {
    batchSize: 100,
    batchDelay: 50,
  },
}: OptimizedGraphProps) => {
  const nodeCount = initialNodes.length;
  const shouldCluster = enableClustering && nodeCount > 100;
  const shouldLoadProgressively = enableProgressiveLoading && nodeCount > 200;
  const shouldUseLODRendering = enableLOD && shouldUseLOD(nodeCount);

  // Apply LOD node type if needed
  const lodNodes = useMemo(() => {
    if (!shouldUseLODRendering) {
      return initialNodes.map(node => ({
        ...node,
        type: node.type === 'cluster' ? 'cluster' : 'custom',
      }));
    }

    return initialNodes.map(node => ({
      ...node,
      type: node.type === 'cluster' ? 'cluster' : 'lod',
    }));
  }, [initialNodes, shouldUseLODRendering]);

  // Apply clustering if needed
  const {
    visibleNodes: clusteredNodes,
    visibleEdges: clusteredEdges,
    clusterMap,
    toggleCluster,
    expandAll,
    collapseAll,
  } = useGraphClustering(
    shouldCluster ? lodNodes : lodNodes,
    initialEdges,
    shouldCluster ? clusteringOptions : { groupBy: 'type', minClusterSize: 999999 }
  );

  // Apply progressive loading if needed
  const {
    visibleNodes: progressiveNodes,
    visibleEdges: progressiveEdges,
    isLoading,
    progress,
    skipToEnd,
  } = useProgressiveGraph(
    shouldLoadProgressively ? clusteredNodes : clusteredNodes,
    shouldLoadProgressively ? clusteredEdges : clusteredEdges,
    shouldLoadProgressively ? progressiveOptions : { batchSize: 999999 }
  );

  // Final nodes and edges to render
  const finalNodes = shouldLoadProgressively ? progressiveNodes : clusteredNodes;
  const finalEdges = shouldLoadProgressively ? progressiveEdges : clusteredEdges;

  // Debounced nodes/edges for performance
  const debouncedNodes = useDebounce(finalNodes, 100);
  const debouncedEdges = useDebounce(finalEdges, 100);

  // ReactFlow state
  const [nodes, , onNodesChange] = useNodesState(debouncedNodes);
  const [edges, , onEdgesChange] = useEdgesState(debouncedEdges);

  // Handle node click with cluster expansion
  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      // If it's a cluster node, toggle expansion
      if (node.type === 'cluster') {
        toggleCluster(node.id);
        return;
      }

      // Otherwise, call the provided handler
      if (onNodeClick) {
        onNodeClick(node);
      }
    },
    [toggleCluster, onNodeClick]
  );

  // Handle edge click
  const handleEdgeClick = useCallback(
    (_: React.MouseEvent, edge: Edge) => {
      if (onEdgeClick) {
        onEdgeClick(edge);
      }
    },
    [onEdgeClick]
  );

  // Show cluster controls if clustering is active
  const [showControls, setShowControls] = useState(true);

  return (
    <div className="relative w-full h-full">
      {/* Progressive loading indicator */}
      {isLoading && shouldLoadProgressively && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-10 bg-white rounded-lg shadow-lg px-4 py-2 flex items-center gap-3">
          <div className="text-sm font-medium text-gray-700">
            Loading graph: {progress}%
          </div>
          <div className="w-32 h-2 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
          <button
            onClick={skipToEnd}
            className="text-xs text-blue-600 hover:text-blue-800 font-medium"
          >
            Skip
          </button>
        </div>
      )}

      {/* Cluster controls */}
      {shouldCluster && showControls && (
        <div className="absolute top-4 right-4 z-10 bg-white rounded-lg shadow-lg p-3 flex flex-col gap-2">
          <div className="text-xs font-bold text-gray-700 mb-1">
            Cluster Controls
          </div>
          <button
            onClick={expandAll}
            className="px-3 py-1.5 text-xs font-medium bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            Expand All
          </button>
          <button
            onClick={collapseAll}
            className="px-3 py-1.5 text-xs font-medium bg-gray-500 text-white rounded hover:bg-gray-600 transition-colors"
          >
            Collapse All
          </button>
          <div className="text-[10px] text-gray-500 mt-1">
            {Array.from(clusterMap.keys()).length} clusters
          </div>
          <button
            onClick={() => setShowControls(false)}
            className="text-[10px] text-gray-400 hover:text-gray-600"
          >
            Hide
          </button>
        </div>
      )}

      {!showControls && shouldCluster && (
        <button
          onClick={() => setShowControls(true)}
          className="absolute top-4 right-4 z-10 px-2 py-1 bg-white rounded-lg shadow-lg text-xs text-gray-600 hover:text-gray-800"
        >
          Show Controls
        </button>
      )}

      {/* Performance stats */}
      <div className="absolute bottom-4 left-4 z-10 bg-white/90 rounded-lg shadow-md px-3 py-2 text-xs text-gray-600">
        <div>Nodes: {finalNodes.length} / {initialNodes.length}</div>
        <div>Edges: {finalEdges.length} / {initialEdges.length}</div>
        {shouldUseLODRendering && <div className="text-green-600 font-medium">LOD: ON</div>}
        {shouldCluster && <div className="text-blue-600 font-medium">Clustering: ON</div>}
        {shouldLoadProgressively && <div className="text-purple-600 font-medium">Progressive: ON</div>}
      </div>

      {/* ReactFlow */}
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange as OnNodesChange}
        onEdgesChange={onEdgesChange as OnEdgesChange}
        onNodeClick={handleNodeClick}
        onEdgeClick={handleEdgeClick}
        nodeTypes={nodeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        fitView
        fitViewOptions={{
          padding: 0.2,
          maxZoom: 1.5,
        }}
        minZoom={0.1}
        maxZoom={2}
        attributionPosition="bottom-right"
        proOptions={{ hideAttribution: true }}
      >
        <Background gap={16} size={1} color="#e2e8f0" />
        <Controls showInteractive={false} />
        <MiniMap
          nodeColor={(node) => {
            if (node.type === 'cluster') return '#9ca3af'; // gray-400
            const severity = node.data?.severity;
            switch (severity) {
              case 'critical': return '#ef4444'; // red-500
              case 'high': return '#f97316'; // orange-500
              case 'medium': return '#eab308'; // yellow-500
              case 'low': return '#3b82f6'; // blue-500
              default: return '#6b7280'; // gray-500
            }
          }}
          maskColor="rgba(0, 0, 0, 0.05)"
          style={{
            backgroundColor: '#f8fafc',
          }}
        />
      </ReactFlow>
    </div>
  );
});

OptimizedGraph.displayName = 'OptimizedGraph';
